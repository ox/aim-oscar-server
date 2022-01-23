package migrations

import (
	"aim-oscar/models"
	"context"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		db.RegisterModel((*models.User)(nil), (*models.Message)(nil), (*models.Buddy)(nil), (*models.EmailVerification)(nil))

		fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
		return fixture.Load(context.Background(), os.DirFS("./"), "init_fixtures.yml")
	}, nil)
}
