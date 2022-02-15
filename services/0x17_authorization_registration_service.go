package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"

	"aim-oscar/models"
	"aim-oscar/oscar"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

const CIPHER_LENGTH = 64
const AIM_MD5_STRING = "AOL Instant Messenger (SM)"

var ROAST = [16]byte{0xF3, 0x26, 0x81, 0xC4, 0x39, 0x86, 0xDB, 0x92, 0x71, 0xA3, 0xB9, 0xE6, 0x53, 0x7A, 0x95, 0x7C}

func roast(password string) []byte {
	ret := make([]byte, 0)
	for i, letter := range password {
		ret = append(ret, byte(letter)^ROAST[i%16])
	}
	return ret
}

type AuthorizationCookie struct {
	UIN int64
	X   string
}

type AuthorizationRegistrationService struct {
	BOSAddress string
}

func AuthenticateFLAPCookie(ctx context.Context, db *bun.DB, flap *oscar.FLAP) (*models.User, error) {
	// Otherwise this is a protocol negotiation from the client. They're likely trying to connect
	// and sending a cookie to verify who they are.
	tlvs, err := oscar.UnmarshalTLVs(flap.Data.Bytes()[4:])
	if err != nil {
		return nil, errors.Wrap(err, "authentication request missing TLVs")
	}

	/*
		There are 2 ways that clients authenticate: channel 1 auth w/ roasted password, or via MD5 hash. The
		former is used by the 1.0 client, whereas the second is used by 3.5 and up (I believe).
	*/

	// This is channel 1 auth
	screenNameTLV := oscar.FindTLV(tlvs, 0x1)
	roastedPWTLV := oscar.FindTLV(tlvs, 0x2)
	if screenNameTLV != nil && roastedPWTLV != nil {
		user, err := models.UserByScreenName(ctx, db, string(screenNameTLV.Data))
		if err != nil {
			return nil, errors.Wrap(err, "could not get User by UIN")
		}

		if !bytes.Equal(roastedPWTLV.Data, roast(user.Password)) {
			return nil, errors.New("invalid password")
		}

		return user, nil
	}

	// This is MD5 hash auth
	cookieTLV := oscar.FindTLV(tlvs, 0x6)
	if cookieTLV == nil {
		return nil, errors.New("authentication request missing Cookie TLV 0x6")
	}

	auth := AuthorizationCookie{}
	if err := json.Unmarshal(cookieTLV.Data, &auth); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal cookie")
	}

	user, err := models.UserByUIN(ctx, db, auth.UIN)
	if err != nil {
		return nil, errors.Wrap(err, "could not get User by UIN")
	}

	h := md5.New()
	io.WriteString(h, user.Cipher)
	io.WriteString(h, user.Password)
	io.WriteString(h, AIM_MD5_STRING)
	expectedPasswordHash := fmt.Sprintf("%x", h.Sum(nil))

	// Make sure the hash passed in matches the one from the DB
	if expectedPasswordHash != auth.X {
		return nil, errors.New("unexpected cookie hash")
	}

	return user, nil
}

func (a *AuthorizationRegistrationService) GenerateCipher() (string, error) {
	randomBytes := make([]byte, 64)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errors.Wrap(err, "could not generate cipher")
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:CIPHER_LENGTH], nil
}

func (a *AuthorizationRegistrationService) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	session, err := oscar.SessionFromContext(ctx)
	if err != nil {
		return ctx, errors.Wrap(err, "could not extract session from context")
	}

	switch snac.Header.Subtype {
	// Request MD5 Auth Key
	case 0x06:
		tlvs, err := oscar.UnmarshalTLVs(snac.Data.Bytes())
		if err != nil {
			return ctx, errors.Wrap(err, "could not unmarshal TLVs")
		}

		screenNameTLV := oscar.FindTLV(tlvs, 1)
		if screenNameTLV == nil {
			return ctx, errors.New("missing screen_name TLV")
		}

		// Fetch the user
		user, err := models.UserByScreenName(ctx, db, string(screenNameTLV.Data))
		if err != nil {
			return ctx, err
		}
		if user == nil {
			snac := oscar.NewSNAC(0x17, 0x03)
			snac.Data.WriteBinary(screenNameTLV)
			snac.Data.WriteBinary(oscar.NewTLV(0x08, []byte{0, 4}))
			resp := oscar.NewFLAP(2)
			resp.Data.WriteBinary(snac)
			return ctx, session.Send(resp)
		}

		// Create cipher for this user
		user.Cipher, err = a.GenerateCipher()
		if err != nil {
			return ctx, err
		}
		if err = user.Update(ctx, db, "cipher"); err != nil {
			return ctx, err
		}

		snac := oscar.NewSNAC(0x17, 0x07)
		snac.Data.WriteUint16(uint16(len(user.Cipher)))
		snac.Data.WriteString(user.Cipher)

		resp := oscar.NewFLAP(2)
		resp.Data.WriteBinary(snac)
		return ctx, session.Send(resp)

	// Client Authorization Request
	case 0x02:
		tlvs, err := oscar.UnmarshalTLVs(snac.Data.Bytes())
		if err != nil {
			return ctx, errors.Wrap(err, "could not unmarshal TLVs")
		}

		screenNameTLV := oscar.FindTLV(tlvs, 1)
		if screenNameTLV == nil {
			return ctx, errors.New("missing screen_name TLV 0x1")
		}

		screen_name := string(screenNameTLV.Data)
		ctx := context.Background()
		user, err := models.UserByScreenName(ctx, db, screen_name)
		if err != nil {
			return ctx, err
		}

		if user == nil {
			snac := oscar.NewSNAC(0x17, 0x03)
			snac.Data.WriteBinary(screenNameTLV)
			snac.Data.WriteBinary(oscar.NewTLV(0x08, []byte{0, 4}))
			resp := oscar.NewFLAP(2)
			resp.Data.WriteBinary(snac)
			return ctx, session.Send(resp)
		}

		passwordHashTLV := oscar.FindTLV(tlvs, 0x25)
		if passwordHashTLV == nil {
			return ctx, errors.New("missing password hash TLV 0x25")
		}

		// Compute password has that we expect the client to send back if the password was right
		h := md5.New()
		io.WriteString(h, user.Cipher)
		io.WriteString(h, user.Password)
		io.WriteString(h, AIM_MD5_STRING)
		expectedPasswordHash := h.Sum(nil)

		if !bytes.Equal(expectedPasswordHash, passwordHashTLV.Data) {
			// Tell the client this was a bad password
			badPasswordSnac := oscar.NewSNAC(0x17, 0x03)
			badPasswordSnac.Data.WriteBinary(screenNameTLV)
			badPasswordSnac.Data.WriteBinary(oscar.NewTLV(0x08, []byte{0, 4})) // incorrect nick/pass
			badPasswordFlap := oscar.NewFLAP(2)
			badPasswordFlap.Data.WriteBinary(badPasswordSnac)
			session.Send(badPasswordFlap)

			// Tell them to leave
			discoFlap := oscar.NewFLAP(4)
			return ctx, session.Send(discoFlap)
		}

		// Only users that have verified their email can use the service
		if !user.Verified || user.DeletedAt != nil {
			// Tell the client this was a bad password
			badPasswordSnac := oscar.NewSNAC(0x17, 0x03)
			badPasswordSnac.Data.WriteBinary(screenNameTLV)
			badPasswordSnac.Data.WriteBinary(oscar.NewTLV(0x08, []byte{0, 7})) // invalid account
			badPasswordSnac.Data.WriteBinary(oscar.NewTLV(0x04, []byte("http://runningman.network/errors/unverified-account")))
			badPasswordFlap := oscar.NewFLAP(2)
			badPasswordFlap.Data.WriteBinary(badPasswordSnac)
			session.Send(badPasswordFlap)

			// Tell them to leave
			discoFlap := oscar.NewFLAP(4)
			return ctx, session.Send(discoFlap)
		}

		// Send BOS response + cookie
		authSnac := oscar.NewSNAC(0x17, 0x3)
		authSnac.Data.WriteBinary(screenNameTLV)
		authSnac.Data.WriteBinary(oscar.NewTLV(0x5, []byte(a.BOSAddress)))

		cookie, err := json.Marshal(AuthorizationCookie{
			UIN: user.UIN,
			X:   fmt.Sprintf("%x", expectedPasswordHash),
		})
		if err != nil {
			return ctx, errors.Wrap(err, "could not marshal authorization cookie")
		}

		authSnac.Data.WriteBinary(oscar.NewTLV(0x6, cookie))
		authSnac.Data.WriteBinary(oscar.NewTLV(0x11, []byte(user.Email)))
		authFlap := oscar.NewFLAP(2)
		authFlap.Data.WriteBinary(authSnac)
		session.Send(authFlap)

		// Tell them to leave
		discoFlap := oscar.NewFLAP(4)
		session.Send(discoFlap)
		return ctx, session.Disconnect()
	}

	return ctx, nil
}
