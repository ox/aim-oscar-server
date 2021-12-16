package main

import "github.com/uptrace/bun"

type Service interface {
	HandleSNAC(*bun.DB, *Session, *SNAC) error
}
