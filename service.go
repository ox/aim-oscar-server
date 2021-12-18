package main

import (
	"aim-oscar/oscar"

	"github.com/uptrace/bun"
)

type Service interface {
	HandleSNAC(*bun.DB, *oscar.Session, *oscar.SNAC) error
}
