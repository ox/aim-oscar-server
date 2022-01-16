package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/util"
	"context"
	"log"
	"time"

	"github.com/uptrace/bun"
)

func OnlineNotification(sm *SessionManager) (chan *models.User, routineFn) {
	commCh := make(chan *models.User, 1)

	routine := func(db *bun.DB) {
		log.Printf("online notification starting up")

		for {
			user, more := <-commCh
			if !more {
				log.Printf("online notification routine shutting down")
				return
			}

			if user.Status == models.UserStatusOnline {
				log.Printf("%s is now online", user.ScreenName)
			} else if user.Status == models.UserStatusAway {
				log.Printf("%s is now away", user.ScreenName)
			}

			ctx := context.Background()

			// Find buddies who are friends with the user
			var buddies []*models.Buddy
			err := db.NewSelect().Model(&buddies).Where("with_uin = ?", user.UIN).Relation("Source").Scan(ctx, &buddies)
			if err != nil {
				log.Printf("could not find user's buddies: %s", err.Error())
				return
			}

			for _, buddy := range buddies {
				if buddy.Source.Status == models.UserStatusAway || buddy.Source.Status == models.UserStatusDnd {
					continue
				}
				log.Printf("telling %s that %s has a new status: %d!", buddy.Source.ScreenName, user.ScreenName, user.Status)

				if s := sm.GetSession(buddy.Source.ScreenName); s != nil {
					onlineSnac := oscar.NewSNAC(3, 0xb)
					onlineSnac.Data.WriteLPString(user.ScreenName)
					onlineSnac.Data.WriteUint16(0) // TODO: user warning level

					tlvs := []*oscar.TLV{
						oscar.NewTLV(1, util.Word(0)),                       // TODO: user class
						oscar.NewTLV(0x06, util.Dword(uint32(user.Status))), // TODO: User Status
						// oscar.NewTLV(0x0a, util.Dword(binary.BigEndian.Uint32([]byte(OSCAR_HOST)))),       // TODO: External IP of the client?
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
						log.Printf("could not tell %s that %s is online", buddy.Source.ScreenName, user.ScreenName)
					}
				}
			}
		}
	}

	return commCh, routine
}
