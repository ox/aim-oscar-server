package main

import (
	"errors"
)

type AuthorizationRegistrationService struct{}

func (a *AuthorizationRegistrationService) HandleSNAC(session *Session, snac *SNAC) error {
	switch snac.Header.Subtype {
	// Request MD5 Auth Key
	case 0x06:
		tlvs, err := UnmarshalTLVs(snac.Data)
		panicIfError(err)

		usernameTLV := FindTLV(tlvs, 1)
		if usernameTLV == nil {
			return errors.New("missing username TLV")
		}

		// Create cipher for this user
		cipher := "howdy"
		db.Set("cipher-"+string(usernameTLV.Data), cipher)
		cipherData := []byte(cipher)

		snac := NewSNAC(0x17, 0x07, cipherData)
		snacBytes, err := snac.MarshalBinary()
		panicIfError(err)

		resp := NewFLAP(session, 2, snacBytes)
		return session.Send(resp)
	}

	return nil
}
