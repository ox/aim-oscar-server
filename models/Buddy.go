package models

import "github.com/uptrace/bun"

type Buddy struct {
	bun.BaseModel `bun:"table:buddies"`
	ID            int   `bun:",pk"`
	SourceUIN     int64 `bun:",notnull"`
	Source        *User `bun:"rel:has-one,join:source_uin=uin"`
	WithUIN       int64 `bun:",notnull"`
	Target        *User `bun:"rel:has-one,join:with_uin=uin"`
}
