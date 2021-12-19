package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/util"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/uptrace/bun"
)

func OnlineNotification() (chan *models.User, routineFn) {
	commCh := make(chan *models.User, 1)

	routine := func(db *bun.DB) {
		log.Printf("online notification starting up")

		for {
			user, more := <-commCh
			if !more {
				log.Printf("online notification routine shutting down")
				return
			}

			if user.Status == models.UserStatusActive {
				fmt.Printf("%s is online", user.Username)

				var buddies []*models.Buddy
				err := db.NewSelect().Model(&buddies).Where("with_uin = ?", user.UIN).Relation("Source").Relation("Target").Scan(context.Background(), &buddies)
				if err != nil {
					log.Printf("could not find user's buddies: %s", err.Error())
					return
				}

				for _, buddy := range buddies {
					if s := getSession(buddy.Source.Username); s != nil {
						onlineSnac := oscar.NewSNAC(3, 0xb)
						onlineSnac.Data.WriteLPString(user.Username)
						onlineSnac.Data.WriteUint16(0) // TODO: user warning level

						tlvs := []*oscar.TLV{
							oscar.NewTLV(1, util.Word(0x80)),                                                  // TODO: user class
							oscar.NewTLV(0x06, util.Dword(0x0001|0x0100)),                                     // TODO: User Status
							oscar.NewTLV(0x0a, util.Dword(binary.BigEndian.Uint32([]byte(SRV_HOST)))),         // External IP
							oscar.NewTLV(0x0f, util.Dword(uint32(time.Since(user.LastActivityAt).Seconds()))), // Idle Time
							oscar.NewTLV(0x03, util.Dword(uint32(time.Now().Unix()))),                         // Client Signon Time
							oscar.NewTLV(0x05, util.Dword(uint32(user.CreatedAt.Unix()))),                     // Member since
						}

						onlineSnac.Data.WriteUint16(uint16(len(tlvs)))
						for _, tlv := range tlvs {
							onlineSnac.Data.WriteBinary(tlv)
						}

						onlineFlap := oscar.NewFLAP(2)
						onlineFlap.Data.WriteBinary(onlineSnac)
						if err := s.Send(onlineFlap); err != nil {
							log.Printf("could not tell %s that %s is online", buddy.Source.Username, buddy.Target.Username)
						}
					}
				}
			}

			if user.Status == models.UserStatusInactive {
				var buddies []*models.Buddy
				err := db.NewSelect().Model(&buddies).Where("with_uin = ?", user.UIN).Relation("Source").Relation("Target").Scan(context.Background(), &buddies)
				if err != nil {
					log.Printf("could not find user's buddies: %s", err.Error())
					return
				}

				for _, buddy := range buddies {
					if s := getSession(buddy.Source.Username); s != nil {
						offlineSnac := oscar.NewSNAC(3, 0xb)
						offlineSnac.Data.WriteLPString(user.Username)
						offlineSnac.Data.WriteUint16(0) // TODO: user warning level
						offlineSnac.Data.WriteUint16(1)
						offlineSnac.Data.WriteBinary(oscar.NewTLV(1, util.Dword(0x80)))

						offlineFlap := oscar.NewFLAP(2)
						offlineFlap.Data.WriteBinary(offlineSnac)
						if err := s.Send(offlineFlap); err != nil {
							log.Printf("could not tell %s that %s is offline", buddy.Source.Username, buddy.Target.Username)
						}
					}
				}
			}
		}
	}

	return commCh, routine
}
