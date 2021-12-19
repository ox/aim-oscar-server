package main

import (
	"aim-oscar/aimerror"
	"aim-oscar/models"
	"aim-oscar/oscar"
	"bytes"
	"context"
	"encoding/binary"
	"log"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type ICBM struct {
	CommCh chan *models.Message
}

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

	// Client wants to send a message to someone through the server
	case 0x06:
		user := models.UserFromContext(ctx)
		if user == nil {
			return ctx, aimerror.NoUserInSession
		}

		msgID, _ := snac.Data.ReadUint64()
		msgChannel, _ := snac.Data.ReadUint16()
		to, _ := snac.Data.ReadLPString()

		if msgChannel != 1 {
			log.Printf("Message for unsupported channel %d", msgChannel)
			return ctx, nil
		}

		tlvs, err := oscar.UnmarshalTLVs(snac.Data.Bytes())
		if err != nil {
			return ctx, errors.Wrap(err, "could not unmarshal message tlvs")
		}

		messageTLV := oscar.FindTLV(tlvs, 0x2)
		if messageTLV == nil {
			return ctx, errors.New("missing messageTLV 0x2")
		}

		// Parse fragment (array of required capabilities, yawn)
		messageTLVData := oscar.Buffer{}
		messageTLVData.Write(messageTLV.Data)

		fragmentNum, err := messageTLVData.ReadUint8()
		if err != nil {
			return ctx, errors.Wrap(err, "could not read fragment identifier")
		} else if fragmentNum != 5 {
			return ctx, errors.New("expected first fragment identifier to be 5")
		}

		fragmentVersion, err := messageTLVData.ReadUint8()
		if err != nil {
			return ctx, errors.Wrap(err, "could not read fragment version")
		} else if fragmentVersion != 1 {
			return ctx, errors.New("expected first fragment version to be 1")
		}

		fragmentLength, err := messageTLVData.ReadUint16()
		if err != nil {
			return ctx, errors.Wrap(err, "could not read fragment data length")
		}

		// Skip over all the capabilities
		messageTLVData.Seek(int(fragmentLength))

		// This should be the start of the message contents fragment
		fragmentNum, err = messageTLVData.ReadUint8()
		if err != nil {
			return ctx, errors.Wrap(err, "could not read fragment identifier")
		} else if fragmentNum != 1 {
			return ctx, errors.New("expected second fragment identifier to be 1")
		}

		fragmentVersion, err = messageTLVData.ReadUint8()
		if err != nil {
			return ctx, errors.Wrap(err, "could not read fragment version")
		} else if fragmentVersion != 1 {
			return ctx, errors.New("expected second fragment version to be 1")
		}

		fragmentLength, err = messageTLVData.ReadUint16()
		if err != nil {
			return ctx, errors.Wrap(err, "could not read second fragment data length")
		}

		// Skip over the charset + language
		messageTLVData.Seek(4)

		messageContents := make([]byte, fragmentLength-4)
		n, err := messageTLVData.Read(messageContents)
		if err != nil {
			return ctx, errors.Wrap(err, "could not read message contents from fragment")
		}
		if n < int(fragmentLength)-4 {
			return ctx, errors.New("read insufficient data from message fragment")
		}

		message, err := models.InsertMessage(ctx, db, msgID, user.Username, to, string(messageContents))
		if err != nil {
			return ctx, errors.Wrap(err, "could not insert message")
		}

		// Fire the message off into the communication channel to get delivered
		icbm.CommCh <- message

		// The Client usually wants a response that the server got the message. It checks that the message
		// back has the same message ID that was sent and the user it was sent to.
		ackTLV := oscar.FindTLV(tlvs, 3)
		if ackTLV != nil {
			ackSnac := oscar.NewSNAC(4, 0xc)
			ackSnac.Data.WriteUint64(msgID)
			ackSnac.Data.WriteUint16(2)
			ackSnac.Data.WriteLPString(user.Username)
			ackFlap := oscar.NewFLAP(2)
			ackFlap.Data.WriteBinary(ackSnac)
			return ctx, session.Send(ackFlap)
		}

		return ctx, nil
	}

	return ctx, nil
}
