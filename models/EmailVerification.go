package models

import (
	"time"

	"github.com/uptrace/bun"
)

// TODO: move this out of here and into API server
type EmailVerification struct {
	bun.BaseModel `bun:"table:email_verification"`
	UserUIN       int64     `bun:",pk,notnull,unique"`
	User          *User     `bun:"rel:has-one,join:user_uin=uin"`
	Token         string    `bun:",notnull"`
	Used          bool      `bun:",notnull,default:false"`
	CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}
