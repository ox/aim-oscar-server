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
	"aim-oscar/util"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

const CIPHER_LENGTH = 64
const AIM_MD5_STRING = "AOL Instant Messenger (SM)"

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

func (a *AuthorizationRegistrationService) GenerateCipher() string {
	randomBytes := make([]byte, 64)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:CIPHER_LENGTH]
}

func (a *AuthorizationRegistrationService) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	session, err := oscar.SessionFromContext(ctx)
	if err != nil {
		util.PanicIfError(err)
	}

	switch snac.Header.Subtype {
	// Request MD5 Auth Key
	case 0x06:
		tlvs, err := oscar.UnmarshalTLVs(snac.Data.Bytes())
		util.PanicIfError(err)

		usernameTLV := oscar.FindTLV(tlvs, 1)
		if usernameTLV == nil {
			return ctx, errors.New("missing username TLV")
		}

		// Fetch the user
		user, err := models.UserByUsername(ctx, db, string(usernameTLV.Data))
		if err != nil {
			return ctx, err
		}
		if user == nil {
			snac := oscar.NewSNAC(0x17, 0x03)
			snac.Data.WriteBinary(usernameTLV)
			snac.Data.WriteBinary(oscar.NewTLV(0x08, []byte{0, 4}))
			resp := oscar.NewFLAP(2)
			resp.Data.WriteBinary(snac)
			return ctx, session.Send(resp)
		}

		// Create cipher for this user
		user.Cipher = a.GenerateCipher()
		if err = user.Update(ctx, db); err != nil {
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
		util.PanicIfError(err)

		usernameTLV := oscar.FindTLV(tlvs, 1)
		if usernameTLV == nil {
			return ctx, errors.New("missing username TLV 0x1")
		}

		username := string(usernameTLV.Data)
		ctx := context.Background()
		user, err := models.UserByUsername(ctx, db, username)
		if err != nil {
			return ctx, err
		}

		if user == nil {
			snac := oscar.NewSNAC(0x17, 0x03)
			snac.Data.WriteBinary(usernameTLV)
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
			badPasswordSnac.Data.WriteBinary(usernameTLV)
			badPasswordSnac.Data.WriteBinary(oscar.NewTLV(0x08, []byte{0, 4}))
			badPasswordFlap := oscar.NewFLAP(2)
			badPasswordFlap.Data.WriteBinary(badPasswordSnac)
			session.Send(badPasswordFlap)

			// Tell them to leave
			discoFlap := oscar.NewFLAP(4)
			return ctx, session.Send(discoFlap)
		}

		// Send BOS response + cookie
		authSnac := oscar.NewSNAC(0x17, 0x3)
		authSnac.Data.WriteBinary(usernameTLV)
		authSnac.Data.WriteBinary(oscar.NewTLV(0x5, []byte(a.BOSAddress)))

		cookie, err := json.Marshal(AuthorizationCookie{
			UIN: user.UIN,
			X:   fmt.Sprintf("%x", expectedPasswordHash),
		})
		util.PanicIfError(err)

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
