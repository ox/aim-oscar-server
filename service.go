package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
	"context"

	"github.com/uptrace/bun"
)

type Service interface {
	HandleSNAC(context.Context, *bun.DB, *oscar.SNAC, chan *models.Message) (context.Context, error)
}
