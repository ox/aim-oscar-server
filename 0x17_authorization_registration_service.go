package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base32"
	"io"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"

	"aim-oscar/models"
)

const CIPHER_LENGTH = 64
const AIM_MD5_STRING = "AOL Instant Messenger (SM)"

type AuthorizationRegistrationService struct{}

func (a *AuthorizationRegistrationService) GenerateCipher() string {
	randomBytes := make([]byte, 64)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:CIPHER_LENGTH]
}

func (a *AuthorizationRegistrationService) HandleSNAC(db *bun.DB, session *Session, snac *SNAC) error {
	switch snac.Header.Subtype {
	// Request MD5 Auth Key
	case 0x06:
		tlvs, err := UnmarshalTLVs(snac.Data.Bytes())
		panicIfError(err)

		usernameTLV := FindTLV(tlvs, 1)
		if usernameTLV == nil {
			return errors.New("missing username TLV")
		}

		// Fetch the user
		ctx := context.Background()
		user, err := models.UserByUsername(ctx, db, string(usernameTLV.Data))
		if err != nil {
			return err
		}
		if user == nil {
			snac := NewSNAC(0x17, 0x03)
			snac.Data.WriteBinary(usernameTLV)
			snac.Data.WriteBinary(NewTLV(0x08, []byte{0, 4}))
			resp := NewFLAP(session, 2)
			resp.Data.WriteBinary(snac)
			return session.Send(resp)
		}

		// Create cipher for this user
		user.Cipher = a.GenerateCipher()
		if err = user.Update(ctx, db); err != nil {
			return err
		}

		snac := NewSNAC(0x17, 0x07)
		snac.Data.WriteUint16(uint16(len(user.Cipher)))
		snac.Data.WriteString(user.Cipher)

		resp := NewFLAP(session, 2)
		resp.Data.WriteBinary(snac)
		return session.Send(resp)

	// Client Authorization Request
	case 0x02:
		tlvs, err := UnmarshalTLVs(snac.Data.Bytes())
		panicIfError(err)

		usernameTLV := FindTLV(tlvs, 1)
		if usernameTLV == nil {
			return errors.New("missing username TLV 0x1")
		}

		username := string(usernameTLV.Data)
		ctx := context.Background()
		user, err := models.UserByUsername(ctx, db, username)
		if err != nil {
			return err
		}

		if user == nil {
			snac := NewSNAC(0x17, 0x03)
			snac.Data.WriteBinary(usernameTLV)
			snac.Data.WriteBinary(NewTLV(0x08, []byte{0, 4}))
			resp := NewFLAP(session, 2)
			resp.Data.WriteBinary(snac)
			return session.Send(resp)
		}

		passwordHashTLV := FindTLV(tlvs, 0x25)
		if passwordHashTLV == nil {
			return errors.New("missing password hash TLV 0x25")
		}

		h := md5.New()
		io.WriteString(h, user.Cipher)
		io.WriteString(h, user.Password)
		io.WriteString(h, AIM_MD5_STRING)
		expectedPasswordHash := h.Sum(nil)

		if !bytes.Equal(expectedPasswordHash, passwordHashTLV.Data) {
			// Tell the client this was a bad password
			badPasswordSnac := NewSNAC(0x17, 0x03)
			badPasswordSnac.Data.WriteBinary(usernameTLV)
			badPasswordSnac.Data.WriteBinary(NewTLV(0x08, []byte{0, 4}))
			badPasswordFlap := NewFLAP(session, 2)
			badPasswordFlap.Data.WriteBinary(badPasswordSnac)
			session.Send(badPasswordFlap)

			// Tell tem to leave
			discoFlap := NewFLAP(session, 4)
			return session.Send(discoFlap)
		}

		// Send BOS response + cookie
		authSnac := NewSNAC(0x17, 0x3)
		authSnac.Data.WriteBinary(usernameTLV)
		authSnac.Data.WriteBinary(NewTLV(0x5, []byte("10.0.1.2:5191")))
		authSnac.Data.WriteBinary(NewTLV(0x6, []byte(`{"hello": "uwu"}`)))
		authFlap := NewFLAP(session, 2)
		authFlap.Data.WriteBinary(authSnac)
		session.Send(authFlap)

		// Tell tem to leave
		discoFlap := NewFLAP(session, 4)
		return session.Send(discoFlap)
	}

	return nil
}
