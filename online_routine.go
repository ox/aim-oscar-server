package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/util"
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"golang.org/x/exp/slog"
)

func OnlineNotification(sm *SessionManager, parentLogger *slog.Logger) (chan *models.User, routineFn) {
	commCh := make(chan *models.User, 1)
	logger := parentLogger.With(slog.String("routine", "online_notification"))

	routine := func(db *bun.DB) {
		logger.Info("Starting up")
		defer logger.Info("Shutting down")

		for {
			user, more := <-commCh
			if !more {
				return
			}

			userLogger := logger.With(slog.String("screen_name", user.ScreenName), slog.String("status", user.Status.String()))
			userLogger.Info("Status change")

			// Find buddies who are friends with the user
			ctx := context.Background()
			var buddies []*models.Buddy
			err := db.NewSelect().Model(&buddies).Where("with_uin = ?", user.UIN).Relation("Source").Scan(ctx, &buddies)
			if err != nil {
				userLogger.Error("Could not find user's buddies", slog.String("err", err.Error()))
				continue
			}

			// Inform each buddy that the user is now online
			for _, buddy := range buddies {
				if buddy.Source.Status == models.UserStatusAway || buddy.Source.Status == models.UserStatusDnd {
					continue
				}
				userLogger.Debug(fmt.Sprintf("notifying %s", buddy.Source.ScreenName))

				if buddySession := sm.GetSession(buddy.Source.ScreenName); buddySession != nil {
					// If the user is now online...
					if user.Status == models.UserStatusOnline {
						onlineSnac := oscar.NewSNAC(0x3, 0xb)
						onlineSnac.Data.WriteLPString(user.ScreenName)
						onlineSnac.Data.WriteUint16(0) // TODO: user warning level

						tlvs := []*oscar.TLV{
							oscar.NewTLV(0x01, util.Word(0x0004)), // TODO: user class
							oscar.NewTLV(0x06, util.Dword(uint32(user.Status))),
							oscar.NewTLV(0x0f, util.Dword(uint32(time.Since(user.LastActivityAt).Seconds()))), // Idle Time
							oscar.NewTLV(0x03, util.Dword(uint32(time.Now().Unix()))),                         // Client Signon Time
							oscar.NewTLV(0x05, util.Dword(uint32(user.CreatedAt.Unix()))),                     // Member since
						}
						onlineSnac.AppendTLVs(tlvs)

						onlineFlap := oscar.NewFLAP(2)
						onlineFlap.Data.WriteBinary(onlineSnac)
						if err := buddySession.Send(onlineFlap); err != nil {
							userLogger.Error(fmt.Sprintf("could not tell %s that %s is online", buddy.Source.ScreenName, user.ScreenName), slog.String("err", err.Error()))
						}

						// If the user is now away
					} else if user.Status == models.UserStatusAway {
						offlineSnac := oscar.NewSNAC(0x3, 0xc)
						offlineSnac.Data.WriteLPString(user.ScreenName)
						offlineSnac.Data.WriteUint16(0) // TODO: user warning level
						tlvs := []*oscar.TLV{
							oscar.NewTLV(1, util.Dword(0x0020)),
						}
						offlineSnac.AppendTLVs(tlvs)

						offlineFlap := oscar.NewFLAP(2)
						offlineFlap.Data.WriteBinary(offlineSnac)
						if err := buddySession.Send(offlineFlap); err != nil {
							userLogger.Error(fmt.Sprintf("could not tell %s that %s is offline", buddy.Source.ScreenName, user.ScreenName), slog.String("err", err.Error()))
						}
					}
				}
			}

			userSession := sm.GetSession(user.ScreenName)
			// If the user is disconnected, don't try to send them notifications
			if userSession == nil {
				continue
			}

			// Get the user's list of online buddies and tell the user that they are online
			for _, buddy := range buddies {
				// If the buddy is away, tell the user
				if buddy.Source.Status == models.UserStatusAway {
					offlineSnac := oscar.NewSNAC(0x3, 0xc)
					offlineSnac.Data.WriteLPString(buddy.Source.ScreenName)
					offlineSnac.Data.WriteUint16(0) // TODO: user warning level
					tlvs := []*oscar.TLV{
						oscar.NewTLV(1, util.Dword(0x0020)),
					}
					offlineSnac.AppendTLVs(tlvs)

					offlineFlap := oscar.NewFLAP(2)
					offlineFlap.Data.WriteBinary(offlineSnac)
					if err := userSession.Send(offlineFlap); err != nil {
						userLogger.Error(fmt.Sprintf("could not tell %s that %s is offline", user.ScreenName, buddy.Source.ScreenName), slog.String("err", err.Error()))
					}
				} else if buddy.Source.Status == models.UserStatusOnline {
					onlineSnac := oscar.NewSNAC(3, 0xb)
					onlineSnac.Data.WriteLPString(buddy.Source.ScreenName)
					onlineSnac.Data.WriteUint16(0) // TODO: user warning level

					tlvs := []*oscar.TLV{
						oscar.NewTLV(0x01, util.Word(0x0004)), // TODO: user class
						oscar.NewTLV(0x06, util.Dword(uint32(buddy.Source.Status))),
						oscar.NewTLV(0x0f, util.Dword(uint32(time.Since(buddy.Source.LastActivityAt).Seconds()))), // Idle Time
						oscar.NewTLV(0x03, util.Dword(uint32(time.Now().Unix()))),                                 // Client Signon Time
						oscar.NewTLV(0x05, util.Dword(uint32(buddy.Source.CreatedAt.Unix()))),                     // Member since
					}
					onlineSnac.AppendTLVs(tlvs)

					onlineFlap := oscar.NewFLAP(2)
					onlineFlap.Data.WriteBinary(onlineSnac)
					if err := userSession.Send(onlineFlap); err != nil {
						userLogger.Error(fmt.Sprintf("could not tell %s that %s is online", user.ScreenName, buddy.Source.ScreenName), slog.String("err", err.Error()))
					}
				}

			}
		}
	}

	return commCh, routine
}
