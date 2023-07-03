package db

import (
	"aim-oscar/config"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func Connect(c *config.DBConfig) (*bun.DB, error) {
	dbaddr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	pgconn := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(dbaddr),
		pgdriver.WithUser(c.User),
		pgdriver.WithPassword(c.Password),
		pgdriver.WithDatabase(c.Name),
		pgdriver.WithInsecure(c.SSLMode == "disable"),
		pgdriver.WithTimeout(5*time.Second),
		pgdriver.WithDialTimeout(5*time.Second),
		pgdriver.WithReadTimeout(5*time.Second),
		pgdriver.WithWriteTimeout(5*time.Second),
	)

	// Set up the DB
	sqldb := sql.OpenDB(pgconn)
	db := bun.NewDB(sqldb, pgdialect.New())
	db.SetConnMaxIdleTime(15 * time.Second)
	db.SetConnMaxLifetime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping db: %w", err)
	}

	return db, nil
}
