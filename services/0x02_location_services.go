package services

import (
	"aim-oscar/aimerror"
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/util"
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type LocationServices struct {
	OnlineCh chan *models.User
}

func (s *LocationServices) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	session, _ := oscar.SessionFromContext(ctx)

	switch snac.Header.Subtype {

	// Client wants to know the limits/permissions for Location services
	case 0x02:
		paramsSnac := oscar.NewSNAC(2, 3)
		paramsSnac.Data.WriteBinary(oscar.NewTLV(1, util.Word(256))) // Max profile length TODO: error if user sends more

		paramsFlap := oscar.NewFLAP(2)
		paramsFlap.Data.WriteBinary(paramsSnac)

		return ctx, session.Send(paramsFlap)

	// Client set profile/away message
	case 0x04:
		user := models.UserFromContext(ctx)
		if user == nil {
			return ctx, aimerror.NoUserInSession
		}

		tlvs, err := oscar.UnmarshalTLVs(snac.Data.Bytes())
		if err != nil {
			return nil, errors.Wrap(err, "authentication request missing TLVs")
		}

		awayMessageTLV := oscar.FindTLV(tlvs, 0x4)
		if awayMessageTLV != nil {
			// Away message encoding is set in TLV 0x3
			awayMessageMimeTLV := oscar.FindTLV(tlvs, 0x3)
			if awayMessageMimeTLV == nil {
				return nil, errors.New("missing away message mime TLV 0x3")
			}
			user.AwayMessage = string(awayMessageTLV.Data)
			user.AwayMessageEncoding = string(awayMessageMimeTLV.Data)
		}

		profileTLV := oscar.FindTLV(tlvs, 0x2)
		if profileTLV != nil {
			profileMimeTLV := oscar.FindTLV(tlvs, 0x1)
			if profileMimeTLV == nil {
				return nil, errors.New("missing away message mime TLV 0x3")
			}
			user.Profile = string(profileTLV.Data)
			user.ProfileEncoding = string(profileMimeTLV.Data)
		}

		if user.AwayMessage == "" {
			user.Status = models.UserStatusOnline
		} else {
			user.Status = models.UserStatusAway
		}

		if err := user.Update(ctx, db, "away_message", "away_message_encoding", "profile", "profile_encoding"); err != nil {
			return ctx, errors.Wrap(err, "could not set away message")
		}

		s.OnlineCh <- user

		return models.NewContextWithUser(ctx, user), nil

	// Client is asking for user information like profile, away message, online state
	case 0x5:
		requestType, err := snac.Data.ReadUint16()
		if err != nil {
			return ctx, errors.Wrap(err, "missing request type")
		}

		requestedScreenName, err := snac.Data.ReadLPString()
		if err != nil {
			return ctx, errors.Wrap(err, "missing requested screen_name")
		}

		requestedUser, err := models.UserByScreenName(ctx, db, requestedScreenName)
		if err != nil {
			return ctx, aimerror.FetchingUser(err, requestedScreenName)
		}
		if requestedUser == nil {
			noMatchSnac := oscar.NewSNAC(0x2, 1)
			noMatchSnac.Data.WriteUint16(0x14) // error code 0x14: No Match
			noMatchFlap := oscar.NewFLAP(2)
			noMatchFlap.Data.WriteBinary(noMatchSnac)
			session.Send(noMatchFlap)
			return ctx, nil
		}

		respSnac := oscar.NewSNAC(2, 6)
		respSnac.Data.WriteLPString(requestedUser.ScreenName)
		respSnac.Data.WriteUint16(0) // TODO: warning level

		idleTime := 0
		if user.LastActivityAt != nil {
			idleTime = time.Since(user.LastActivityAt).Seconds()
		}

		tlvs := []*oscar.TLV{
			oscar.NewTLV(1, util.Dword(0)),                            // user class
			oscar.NewTLV(6, util.Dword(uint32(requestedUser.Status))), // user status
			// oscar.NewTLV(0x0a, util.Dword(binary.BigEndian.Uint32([]byte(OSCAR_HOST)))),                  // user external IP
			oscar.NewTLV(0x0f, util.Dword(uint32(idleTime))), // idle time
			oscar.NewTLV(0x03, util.Dword(uint32(time.Now().Unix()))),                                  // TODO: signon time
			oscar.NewTLV(0x05, util.Dword(uint32(requestedUser.CreatedAt.Unix()))),                     // member since
		}

		respSnac.AppendTLVs(tlvs)

		// General info (Profile)
		if requestType == 1 {
			respSnac.Data.WriteBinary(oscar.NewTLV(1, []byte(requestedUser.ProfileEncoding)))
			respSnac.Data.WriteBinary(oscar.NewTLV(2, []byte(requestedUser.Profile)))
		}

		// Request Type 2 = online status, no TLVs

		// Away message
		if requestType == 3 {
			respSnac.Data.WriteBinary(oscar.NewTLV(3, []byte(requestedUser.AwayMessageEncoding)))
			respSnac.Data.WriteBinary(oscar.NewTLV(4, []byte(requestedUser.AwayMessage)))
		}

		// TODO: Request Type 4 - User capabilities

		respFlap := oscar.NewFLAP(2)
		respFlap.Data.WriteBinary(respSnac)

		return ctx, session.Send(respFlap)

	case 0xb:
		/* Nobody seems to know what this client request is for
		- http://iserverd.khstu.ru/oscar/snac_02_0b.html
		- https://bugs.bitlbee.org/browser/protocols/oscar/info.c?rev=b7d3cc34f68dab7b8f7d0777711317b334fc2219#L572

		But the one dump that exists looks like a TLV 0x1 with empty data
		*/
		unknownSnac := oscar.NewSNAC(2, 0xc)
		unknownSnac.Data.WriteUint16(1)
		unknownSnac.Data.WriteUint16(0)
		unknownFlap := oscar.NewFLAP(2)
		unknownFlap.Data.WriteBinary(unknownSnac)
		return ctx, session.Send(unknownFlap)
	}

	return ctx, nil
}
