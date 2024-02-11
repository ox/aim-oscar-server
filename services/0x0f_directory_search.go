package services

import (
	"aim-oscar/oscar"
	"context"

	"github.com/uptrace/bun"
)

type DirectorySearchService struct{}

func (d *DirectorySearchService) HandleSNAC(ctx context.Context, db *bun.DB, snac *oscar.SNAC) (context.Context, error) {
	return ctx, nil
}
