package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Feedbag struct {
	bun.BaseModel `bun:"table:feedbag"`

	ScreenName   string
	GroupId      int
	ItemId       int
	ClassId      int
	Name         string
	Attributes   []byte
	LastModified time.Time
}
