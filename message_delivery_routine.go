package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/util"
	"context"
	"time"

	"github.com/uptrace/bun"
	"golang.org/x/exp/slog"
)

type routineFn func(db *bun.DB)

func MessageDelivery(sm *SessionManager, parentLogger *slog.Logger) (chan *models.Message, routineFn) {
	commCh := make(chan *models.Message, 1)
	logger := parentLogger.With(slog.String("routine", "message_delivery"))

	routine := func(db *bun.DB) {
		logger.Info("starting up")
		defer logger.Info("shutting down")

		for {
			message, more := <-commCh
			if !more {
				return
			}

			msgLogger := logger.
				With(slog.Group("message", slog.String("from", message.From), slog.String("to", message.To), slog.Uint64("cookie", message.Cookie)))

			// If the user isn't connected, don't send the message
			session := sm.GetSession(message.To)
			if session == nil {
				continue
			}

			messageSnac := oscar.NewSNAC(4, 7)
			messageSnac.Data.WriteUint64(message.Cookie)
			messageSnac.Data.WriteUint16(1)
			messageSnac.Data.WriteLPString(message.From)
			messageSnac.Data.WriteUint16(0) // TODO: sender's warning level

			ctx := context.Background()
			user, err := models.UserByScreenName(ctx, db, message.From)
			if err != nil {
				msgLogger.Error("could not get message author User, can't send message", "err", err.Error())
				continue
			}


			idleTime := 0
			if user.LastActivityAt != nil {
				idleTime = time.Since(user.LastActivityAt).Seconds()
			}

			tlvs := []*oscar.TLV{
				oscar.NewTLV(1, util.Word(0)),                    // TODO: user class
				oscar.NewTLV(6, util.Dword(uint32(user.Status))), // TODO: user status
				oscar.NewTLV(0x0f, util.Dword(uint32(idleTime))), // Idle Time
				oscar.NewTLV(0x03, util.Dword(uint32(idleTime))), // TODO: SignOn time
				// oscar.NewTLV(4, []byte{}), // TODO: this TLV appears in automated responses like away messages
			}

			messageSnac.AppendTLVs(tlvs)

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
			if err := session.Send(messageFlap); err != nil {
				msgLogger.Error("Could not deliver message", slog.String("err", err.Error()))
				continue
			} else {
				msgLogger.Info("Delivered message")
			}

			if message.StoreOffline {
				if err := message.MarkDelivered(context.Background(), db); err != nil {
					msgLogger.Error("could not mark message as delivered", slog.String("err", err.Error()))
				}
			}
		}
	}

	return commCh, routine
}
