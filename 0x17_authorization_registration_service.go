package main

import (
	"context"
	"fmt"
)

var cipher = "hey wassup"

type AuthorizationRegistrationService struct{}

func (a *AuthorizationRegistrationService) HandleSNAC(ctx context.Context, snac *SNAC) {
	// Request MD5 Auth Key
	if snac.Header.Subtype == 0x06 {
		fmt.Println("damn it's 0x06")
		// cipherData := ByteString(cipher) // []byte

		// snac := NewSNAC(0x17, 0x07, cipherData)

		// resp := NewFLAP(2, snac)
	}
}
