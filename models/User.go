package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type UserStatus int16

func (u UserStatus) String() string {
	switch u {
	case UserStatusOnline:
		return "Online"
	case UserStatusAway:
		return "Away"
	case UserStatusDnd:
		return "Dnd"
	case UserStatusNA:
		return "NA"
	case UserStatusOccupied:
		return "Occupied"
	case UserStatusFree4Chat:
		return "Free4Chat"
	case UserStatusInvisible:
		return "Invisible"
	default:
		return "Unknown"
	}
}

func (u UserStatus) Connected() bool {
	switch u {
	case UserStatusOnline:
		return true
	case UserStatusAway:
		return false
	case UserStatusDnd:
		return false
	case UserStatusNA:
		return false
	case UserStatusOccupied:
		return true
	case UserStatusFree4Chat:
		return true
	case UserStatusInvisible:
		return true
	default:
		return false
	}
}

const (
	UserStatusOnline    = 0
	UserStatusAway      = 1
	UserStatusDnd       = 2
	UserStatusNA        = 4
	UserStatusOccupied  = 0x10
	UserStatusFree4Chat = 0x20
	UserStatusInvisible = 0x100
)

type User struct {
	bun.BaseModel       `bun:"table:users"`
	UIN                 int64  `bun:",pk,autoincrement"`
	Email               string `bun:",unique"`
	ScreenName          string `bun:",unique"`
	Password            string
	Cipher              string
	CreatedAt           time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt           time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	DeletedAt           *time.Time `bun:",nullzero"`
	Status              UserStatus
	Verified            bool `bun:",notnull,default:false"`
	Profile             string
	ProfileEncoding     string
	AwayMessage         string
	AwayMessageEncoding string
	LastActivityAt      time.Time `bin:"-"`
}

func (user *User) SetAway(ctx context.Context, db *bun.DB) error {
	user.Status = UserStatusAway
	user.Cipher = ""
	if err := user.Update(ctx, db, "status", "cipher"); err != nil {
		return errors.Wrap(err, "could not set user as inactive")
	}

	return nil
}

type userKey string

func (s userKey) String() string {
	return "user-" + string(s)
}

var (
	currentUser = userKey("user")
)

func CreateUser(ctx context.Context, db *bun.DB, screen_name, password, email string) (*User, error) {
	user := &User{
		ScreenName: screen_name,
		Password:   password,
		Email:      email,
	}

	_, err := db.NewInsert().Model(user).Exec(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "could not create user")
	}

	return user, nil
}

func UserByScreenName(ctx context.Context, db *bun.DB, screen_name string) (*User, error) {
	user := new(User)
	if err := db.NewSelect().Model(user).Where("screen_name = ?", screen_name).Scan(ctx, user); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "could not fetch user")
	}
	return user, nil
}

func UserByUIN(ctx context.Context, db *bun.DB, uin int64) (*User, error) {
	user := new(User)
	if err := db.NewSelect().Model(user).Where("uin = ?", uin).Scan(ctx, user); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "could not fetch user")
	}
	return user, nil
}

func NewContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, currentUser, user)
}

func UserFromContext(ctx context.Context) *User {
	v := ctx.Value(currentUser)
	if v == nil {
		return nil
	}
	return v.(*User)
}

func (u *User) Update(ctx context.Context, db *bun.DB, cols ...string) error {
	q := db.NewUpdate().Model(u).WherePK("uin")

	if len(cols) > 0 {
		q = q.Column(cols...)
	}

	if _, err := q.Exec(ctx); err != nil {
		return errors.Wrap(err, "could not update user")
	}
	return nil
}
