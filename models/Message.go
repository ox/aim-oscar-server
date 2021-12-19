package models

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type Message struct {
	bun.BaseModel `bun:"table:messages"`
	ID            int    `bun:",pk"`
	Cookie        uint64 `bun:",notnull"`
	From          string
	To            string
	Contents      string
	CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	DeliveredAt   time.Time `bun:",nullzero"`
}

func InsertMessage(ctx context.Context, db *bun.DB, cookie uint64, from string, to string, contents string) (*Message, error) {
	msg := &Message{
		Cookie:   cookie,
		From:     from,
		To:       to,
		Contents: contents,
	}
	if _, err := db.NewInsert().Model(msg).Exec(ctx, msg); err != nil {
		return nil, errors.Wrap(err, "could not update user")
	}

	return msg, nil
}

func (m *Message) String() string {
	return fmt.Sprintf("<Message from=%s to=%s content=\"%s\">", m.From, m.To, m.Contents)
}

func (m *Message) MarkDelivered(ctx context.Context, db *bun.DB) error {
	m.DeliveredAt = time.Now()
	if _, err := db.NewUpdate().Model(m).Where("cookie = ?", m.Cookie).Exec(ctx); err != nil {
		return errors.Wrap(err, "could not mark message as updated")
	}

	return nil
}
