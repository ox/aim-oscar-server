package main

import (
	"aim-oscar/oscar"
	"context"

	"github.com/uptrace/bun"
)

type Service interface {
	HandleSNAC(context.Context, *bun.DB, *oscar.SNAC) (context.Context, error)
}
