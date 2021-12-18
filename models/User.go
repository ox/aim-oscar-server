package models

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`
	UIN           int    `bun:",pk,autoincrement"`
	Email         string `bun:",unique"`
	Username      string `bun:",unique"`
	Password      string
	Cipher        string
}

func UserByUsername(ctx context.Context, db *bun.DB, username string) (*User, error) {
	user := new(User)
	if err := db.NewSelect().Model(user).Where("username = ?", username).Scan(ctx, user); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "could not fetch user")
	}
	return user, nil
}

func (u *User) Update(ctx context.Context, db *bun.DB) error {
	if _, err := db.NewUpdate().Model(u).WherePK("uin").Exec(ctx); err != nil {
		return errors.Wrap(err, "could not update user")
	}
	return nil
}
