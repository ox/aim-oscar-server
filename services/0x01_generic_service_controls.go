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

var ServiceVersions map[uint16]uint16

func init() {
	ServiceVersions = make(map[uint16]uint16)
	ServiceVersions[1] = 3
	ServiceVersions[2] = 1
	ServiceVersions[3] = 1
	ServiceVersions[4] = 1
	ServiceVersions[17] = 1
}

type GenericServiceControls struct {
	OnlineCh       chan *models.User
	ServerHostname string
}

func (g *GenericServiceControls) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	session, _ := oscar.SessionFromContext(ctx)

	switch snac.Header.Subtype {

	// Client is ONLINE and READY
	case 0x02:
		user := models.UserFromContext(ctx)
		if user != nil {
			user.Status = models.UserStatusOnline
			if err := user.Update(ctx, db, "status"); err != nil {
				return ctx, errors.Wrap(err, "could not set user as active")
			}

			g.OnlineCh <- user

			return models.NewContextWithUser(ctx, user), nil
		}

		return ctx, nil

	// Client wants to know the rate limits for all services
	case 0x06:
		rateSnac := oscar.NewSNAC(1, 7)
		rateSnac.Data.WriteUint16(1) // one rate class

		// Define a Rate Class
		rc := oscar.Buffer{}
		rc.WriteUint16(1)    // ID
		rc.WriteUint32(80)   // Window Size
		rc.WriteUint32(2500) // Clear level
		rc.WriteUint32(2000) // Alert level
		rc.WriteUint32(1500) // Limit level
		rc.WriteUint32(800)  // Disconnect level
		rc.WriteUint32(3400) // Current level (fake)
		rc.WriteUint32(6000) // Max level
		rc.WriteUint32(0)    // Last time ?
		rc.WriteUint8(0)     // Current state ?
		rateSnac.Data.Write(rc.Bytes())

		// Define a Rate Group
		rg := oscar.Buffer{}
		rg.WriteUint16(1) // ID

		// TODO: make actual rate groups instead of this hack. I can't tell which subtypes are supported so
		// make it set rate limits for everything family for all subtypes under 0x21.
		rg.WriteUint16(uint16(len(ServiceVersions)) * 0x21) // Number of rate groups
		for family := range ServiceVersions {
			for subtype := 0; subtype < 0x21; subtype++ {
				rg.WriteUint16(family)
				rg.WriteUint16(uint16(subtype))
			}
		}
		rateSnac.Data.Write(rg.Bytes())

		rateFlap := oscar.NewFLAP(2)
		rateFlap.Data.WriteBinary(rateSnac)
		return ctx, session.Send(rateFlap)

	// Client wants their own online information
	case 0x0e:
		user := models.UserFromContext(ctx)
		if user == nil {
			return ctx, aimerror.NoUserInSession
		}

		onlineSnac := oscar.NewSNAC(1, 0xf)
		onlineSnac.Data.WriteUint8(uint8(len(user.ScreenName)))
		onlineSnac.Data.WriteString(user.ScreenName)
		onlineSnac.Data.WriteUint16(0) // warning level

		user.Status = models.UserStatusOnline
		if err := user.Update(ctx, db, "status"); err != nil {
			return ctx, errors.Wrap(err, "could not set user as active")
		}

		idleTime := 0
		if user.LastActivityAt != nil {
			idleTime = time.Since(user.LastActivityAt).Seconds()
		}

		tlvs := []*oscar.TLV{
			oscar.NewTLV(0x01, util.Dword(0)),                   // User Class
			oscar.NewTLV(0x06, util.Dword(uint32(user.Status))), // TODO: User Status
			// oscar.NewTLV(0x0a, util.Dword(binary.BigEndian.Uint32([]byte(g.ServerHostname)))), // External IP of the client?
			oscar.NewTLV(0x0f, util.Dword(uint32(idleTime))), // Idle Time
			oscar.NewTLV(0x03, util.Dword(uint32(time.Now().Unix()))),                         // Client Signon Time
			oscar.NewTLV(0x01e, util.Dword(0x0)),                                              // Unknown value
			oscar.NewTLV(0x05, util.Dword(uint32(user.CreatedAt.Unix()))),                     // Member since
		}

		onlineSnac.AppendTLVs(tlvs)

		onlineFlap := oscar.NewFLAP(2)
		onlineFlap.Data.WriteBinary(onlineSnac)
		return models.NewContextWithUser(ctx, user), session.Send(onlineFlap)

	case 0x16:
		// NOP, client keepalive
		return ctx, nil

	// Client wants to know the ServiceVersions of all of the services offered
	case 0x17:
		versionsSnac := oscar.NewSNAC(1, 0x18)
		for family, version := range ServiceVersions {
			versionsSnac.Data.WriteUint16(family)
			versionsSnac.Data.WriteUint16(version)
		}
		versionsFlap := oscar.NewFLAP(2)
		versionsFlap.Data.WriteBinary(versionsSnac)
		return ctx, session.Send(versionsFlap)
	}

	return ctx, nil
}
