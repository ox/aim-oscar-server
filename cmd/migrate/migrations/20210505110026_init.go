package migrations

import (
	"aim-oscar/models"
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		db.RegisterModel((*models.User)(nil), (*models.Message)(nil), (*models.Buddy)(nil), (*models.EmailVerification)(nil), (*models.Feedbag)(nil))

		fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
		return fixture.Load(context.Background(), sqlMigrations, "20210505110026_fixtures.yml")
	}, func(ctx context.Context, db *bun.DB) error {
		models := []interface{}{(*models.User)(nil), (*models.Message)(nil), (*models.Buddy)(nil), (*models.EmailVerification)(nil), (*models.Feedbag)(nil)}

		for _, model := range models {
			if _, err := db.NewDropTable().Model(model).Exec(ctx, nil); err != nil {
				return err
			}
		}

		return nil
	})
}
