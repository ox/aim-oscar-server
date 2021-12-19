package models

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type Message struct {
	bun.BaseModel `bun:"table:messages"`
	MessageID     uint64 `bun:",pk,notnull,unique"`
	From          string
	To            string
	Contents      string
	CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	DeliveredAt   time.Time `bun:",nullzero"`
}

func InsertMessage(ctx context.Context, db *bun.DB, messageId uint64, from string, to string, contents string) error {
	msg := &Message{
		MessageID: messageId,
		From:      from,
		To:        to,
		Contents:  contents,
	}
	if _, err := db.NewInsert().Model(msg).Exec(ctx); err != nil {
		return errors.Wrap(err, "could not update user")
	}

	return nil
}
