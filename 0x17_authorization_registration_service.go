package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"

	"aim-oscar/models"
)

const CIPHER_LENGTH = 64

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
			return errors.New("missing username TLV")
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

		snac := NewSNAC(0x17, 0x03)
		resp := NewFLAP(session, 2)
		resp.Data.WriteBinary(snac)
		return session.Send(resp)
	}

	return nil
}
