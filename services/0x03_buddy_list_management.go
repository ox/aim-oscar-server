package services

import (
	"aim-oscar/aimerror"
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/util"
	"context"
	"log"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type BuddyListManagement struct{}

func (b *BuddyListManagement) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	session, _ := oscar.SessionFromContext(ctx)

	switch snac.Header.Subtype {

	// Client wants to know the buddy list params + limitations
	case 0x2:
		limitSnac := oscar.NewSNAC(3, 3)
		limitSnac.Data.WriteBinary(oscar.NewTLV(1, util.Word(500))) // Max buddy list size
		limitSnac.Data.WriteBinary(oscar.NewTLV(2, util.Word(750))) // Max list watchers
		limitSnac.Data.WriteBinary(oscar.NewTLV(3, util.Word(512))) // Max online notifications ?

		limitFlap := oscar.NewFLAP(2)
		limitFlap.Data.WriteBinary(limitSnac)
		return ctx, session.Send(limitFlap)

	// Add buddy
	case 0x4:
		user := models.UserFromContext(ctx)
		if user == nil {
			return ctx, aimerror.NoUserInSession
		}

		for len(snac.Data.Bytes()) > 0 {
			buddyScreename, err := snac.Data.ReadLPString()
			if err != nil {
				return ctx, errors.Wrap(err, "expecting more buddies in list")
			}

			buddy, err := models.UserByScreenName(ctx, db, buddyScreename)
			if err != nil {
				return ctx, errors.Wrap(err, "error looking for User")
			}
			if buddy == nil {
				noMatchSnac := oscar.NewSNAC(0x3, 1)
				noMatchSnac.Data.WriteUint16(0x14) // error code 0x14: No Match
				noMatchFlap := oscar.NewFLAP(2)
				noMatchFlap.Data.WriteBinary(noMatchSnac)
				session.Send(noMatchFlap)
				return ctx, nil
			}

			rel := &models.Buddy{
				SourceUIN: user.UIN,
				WithUIN:   buddy.UIN,
			}

			_, err = db.NewInsert().Model(rel).Exec(ctx)
			if err != nil {
				return ctx, err
			}

			log.Printf("%s added buddy %s to buddy list", user.ScreenName, buddyScreename)
		}

		return ctx, nil

	// Remove buddies from user list
	case 0x5:
		user := models.UserFromContext(ctx)
		if user == nil {
			return ctx, aimerror.NoUserInSession
		}

		for len(snac.Data.Bytes()) > 0 {
			buddyScreename, err := snac.Data.ReadLPString()
			if err != nil {
				return ctx, errors.Wrap(err, "expecting more buddies in list")
			}

			buddy, err := models.UserByScreenName(ctx, db, buddyScreename)
			if err != nil {
				return ctx, errors.Wrap(err, "error looking for User")
			}
			if buddy == nil {
				noMatchSnac := oscar.NewSNAC(0x3, 1)
				noMatchSnac.Data.WriteUint16(0x14) // error code 0x14: No Match
				noMatchFlap := oscar.NewFLAP(2)
				noMatchFlap.Data.WriteBinary(noMatchSnac)
				session.Send(noMatchFlap)
				return ctx, nil
			}

			_, err = db.NewDelete().Model((*models.Buddy)(nil)).Where("source_uin = ?", user.UIN).Where("with_uin = ?", buddy.UIN).Exec(ctx)
			if err != nil {
				return ctx, err
			}

			log.Printf("%s removed buddy %s from buddy list", user.ScreenName, buddyScreename)
		}

		return ctx, nil
	}

	return ctx, nil
}
