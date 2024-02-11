package services

import (
	"aim-oscar/oscar"
	"context"

	"github.com/uptrace/bun"
)

type AlertService struct{}

// This service doesn't seem to do anything
func (a *AlertService) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	return ctx, nil
}
