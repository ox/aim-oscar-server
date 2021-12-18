package main

import (
	"aim-oscar/oscar"
	"bytes"
	"context"
	"encoding/binary"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type ICBM struct{}

type icbmKey string

func (s icbmKey) String() string {
	return "icbm-" + string(s)
}

var (
	channelKey = icbmKey("channel")
)

func NewContextWithChannel(ctx context.Context, c *channel) context.Context {
	return context.WithValue(ctx, channelKey, c)
}

func ChannelFromContext(ctx context.Context) *channel {
	s := ctx.Value(channelKey)
	if s == nil {
		return nil
	}
	return s.(*channel)
}

type channel struct {
	ID                      uint16
	MessageFlags            uint32
	MaxMessageSnacSize      uint16
	MaxSenderWarningLevel   uint16
	MaxReceiverWarningLevel uint16
	MinimumMessageInterval  uint16
	Unknown                 uint16
}

func (icbm *ICBM) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	session, _ := oscar.SessionFromContext(ctx)

	switch snac.Header.Subtype {
	// Client is telling us about their ICBM capabilities
	case 0x02:
		/*
			xx xx	 	word	 	channel to setup
			xx xx xx xx	 		dword	 	message flags
			xx xx	 	word	 	max message snac size
			xx xx	 	word	 	max sender warning level
			xx xx	 	word	 	max receiver warning level
			xx xx	 	word	 	minimum message interval (sec)
			00 00	 	word	 	unknown parameter (also seen 03 E8)
		*/

		channel := channel{}
		r := bytes.NewReader(snac.Data.Bytes())
		if err := binary.Read(r, binary.BigEndian, &channel); err != nil {
			return ctx, errors.Wrap(err, "could not read channel settings")
		}

		newCtx := NewContextWithChannel(ctx, &channel)
		return newCtx, nil

	// Client asks about the ICBM capabilities we set for them
	case 0x04:
		channel := ChannelFromContext(ctx)
		channelSnac := oscar.NewSNAC(4, 5)
		channelSnac.Data.WriteUint16(uint16(channel.ID))
		channelSnac.Data.WriteUint32(channel.MessageFlags)
		channelSnac.Data.WriteUint16(channel.MaxMessageSnacSize)
		channelSnac.Data.WriteUint16(channel.MaxSenderWarningLevel)
		channelSnac.Data.WriteUint16(channel.MaxReceiverWarningLevel)
		channelSnac.Data.WriteUint16(channel.MinimumMessageInterval)
		channelSnac.Data.WriteUint16(channel.Unknown)

		channelFlap := oscar.NewFLAP(2)
		channelFlap.Data.WriteBinary(channelSnac)
		session.Send(channelFlap)

		return ctx, nil
	}

	return ctx, nil
}
