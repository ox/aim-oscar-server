package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/util"
	"context"
	"log"

	"github.com/uptrace/bun"
)

type routineFn func(db *bun.DB)

func MessageDelivery(sm *SessionManager) (chan *models.Message, routineFn) {
	commCh := make(chan *models.Message, 1)

	routine := func(db *bun.DB) {
		log.Printf("message delivery routine started")

		for {
			message, more := <-commCh
			if !more {
				log.Printf("message delivery routine shutting down")
				return
			}

			log.Printf("got a message: %s", message)
			if s := sm.GetSession(message.To); s != nil {
				messageSnac := oscar.NewSNAC(4, 7)
				messageSnac.Data.WriteUint64(message.Cookie)
				messageSnac.Data.WriteUint16(1)
				messageSnac.Data.WriteLPString(message.From)
				messageSnac.Data.WriteUint16(0) // TODO: sender's warning level

				tlvs := []*oscar.TLV{
					oscar.NewTLV(1, util.Word(0)),    // TODO: user class
					oscar.NewTLV(6, util.Dword(0)),   // TODO: user status
					oscar.NewTLV(0xf, util.Dword(0)), // TODO: user idle time
					oscar.NewTLV(3, util.Dword(0)),   // TODO: user creation time
					// oscar.NewTLV(4, []byte{}), // TODO: this TLV appears in automated responses like away messages
				}

				// Length of TLVs in fixed part
				messageSnac.Data.WriteUint16(uint16(len(tlvs)))

				// Write all of the TLVs to the SNAC
				for _, tlv := range tlvs {
					messageSnac.Data.WriteBinary(tlv)
				}

				frag := oscar.Buffer{}
				frag.Write([]byte{5, 1, 0, 4, 1, 1, 1, 2})          // TODO: first fragment [id, version, len, len, (cap * len)... ]
				frag.Write([]byte{1, 1})                            // message text fragment start (this is a busted "TLV")
				frag.WriteUint16(uint16(len(message.Contents) + 4)) // length of TLV
				frag.Write([]byte{0, 0, 0, 0})                      // TODO: message charset number, message charset subset
				frag.WriteString(message.Contents)

				// Append the fragments
				messageSnac.Data.WriteBinary(oscar.NewTLV(2, frag.Bytes()))

				messageFlap := oscar.NewFLAP(2)
				messageFlap.Data.WriteBinary(messageSnac)
				if err := s.Send(messageFlap); err != nil {
					log.Panicf("could not deliver message %d: %s", message.Cookie, err.Error())
					continue
				} else {
					log.Printf("sent message %d to user %s at %s", message.Cookie, message.To, s.RemoteAddr())
				}

				if message.StoreOffline {
					if err := message.MarkDelivered(context.Background(), db); err != nil {
						log.Panicf("could not mark message %d as delivered: %s", message.Cookie, err.Error())
					}
				}
			} else {
				log.Printf("could not find session for user %s", message.To)
			}
		}
	}

	return commCh, routine
}
