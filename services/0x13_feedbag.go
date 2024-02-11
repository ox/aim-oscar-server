package services

import (
	"aim-oscar/oscar"
	"aim-oscar/util"
	"bytes"
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

type FeedbagService struct{}

type FeedbagItemType uint16

var (
	FeedbagItemTypeUser             FeedbagItemType = 0x0000
	FeedbagItemTypeGroup                            = 0x0001
	FeedbagItemTypePermit                           = 0x0002
	FeedbagItemTypeDeny                             = 0x0003
	FeedbagItemTypePDSetting                        = 0x0004 // permit/deny settings and/or AIM class bitmask
	FeedbagItemTypePresenceInfo                     = 0x0005
	FeedbagItemTypeIgnoreList                       = 0x000e
	FeedbagItemTypeLastUpdateTime                   = 0x000f
	FeedbagItemTypeRosterImportTime                 = 0x0013
	FeedbagItemTypeIconInfo                         = 0x0014 // avatar id
)

type FeedbagItem struct {
	Name           string
	GroupID        uint16
	ItemID         uint16
	ItemType       FeedbagItemType
	AdditionalData []*oscar.TLV
}

func (f *FeedbagItem) Bytes() []byte {
	buf := bytes.Buffer{}

	buf.Write(util.LPUint16String(f.Name))
	buf.Write(util.Word(f.GroupID))
	buf.Write(util.Word(f.ItemID))
	buf.Write(util.Word(uint16(f.ItemType)))
	buf.Write(util.Word(uint16(len(f.AdditionalData))))

	if len(f.AdditionalData) > 0 {
		for _, tlv := range f.AdditionalData {
			b, _ := tlv.MarshalBinary()
			buf.Write(b)
		}
	}

	return buf.Bytes()
}

func (f *FeedbagService) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	session, _ := oscar.SessionFromContext(ctx)
	logger := session.Logger.With("service", "feedbag")

	switch snac.Header.Subtype {

	// Client requests SSI service limitations
	case 0x02:

		respSnac := oscar.NewSNAC(0x13, 0x3)

		maxitems := [][]byte{
			util.Word(0x3D),
			util.Word(0x3D),
			util.Word(0x64),
			util.Word(0x64),
			util.Word(0x01),
			util.Word(0x01),
			util.Word(0x32),
			util.Word(0x00),
			util.Word(0x00),
			util.Word(0x03),
			util.Word(0x00),
			util.Word(0x00),
			util.Word(0x00),
			util.Word(0x80),
			util.Word(0xFF),
			util.Word(0x14),
			util.Word(0xC8),
			util.Word(0x01),
			util.Word(0x00),
			util.Word(0x01),
			util.Word(0x00),
		}
		maxitemsBuf := bytes.Buffer{}
		for _, item := range maxitems {
			maxitemsBuf.Write(item)
		}
		respSnac.WriteTLV(oscar.NewTLV(0x04, maxitemsBuf.Bytes()))
		respSnac.WriteTLV(oscar.NewTLV(0x02, util.Word(0xfe)))
		respSnac.WriteTLV(oscar.NewTLV(0x03, util.Word(0x01fc)))
		respSnac.WriteTLV(oscar.NewTLV(0x05, util.Word(0)))
		respSnac.WriteTLV(oscar.NewTLV(0x06, util.Word(0x61)))
		respSnac.WriteTLV(oscar.NewTLV(0x07, util.Word(0x0a)))

		respFlap := oscar.NewFLAP(2)
		respFlap.Data.WriteBinary(respSnac)

		return ctx, session.Send(respFlap)

	case 0x04:
		respSnac := oscar.NewSNAC(0x13, 0x6)

		respSnac.Data.WriteUint8(0) // SSI Version
		items := []FeedbagItem{}

		respSnac.Data.WriteUint16(uint16(len(items)))
		respSnac.Data.WriteUint32(0) // SSI last change time // TODO: add SSI change time

		respFlap := oscar.NewFLAP(2)
		respFlap.Data.WriteBinary(respSnac)

		return ctx, session.Send(respFlap)
	}

	logger.Error(fmt.Sprintf("Unknown feedbag family/subtype: 0x13, 0x%02x", snac.Header.Subtype))

	return ctx, nil
}
