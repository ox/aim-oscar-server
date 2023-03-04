package main

import (
	"aim-oscar/cmd/migrate/migrations"
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

var (
	DB_URL      = ""
	DB_USER     = ""
	DB_PASSWORD = ""
	DB_NAME     = ""
)

func init() {
	if dbUrl, ok := os.LookupEnv("DB_URL"); ok {
		DB_URL = strings.TrimSpace(dbUrl)
	}

	if dbUser, ok := os.LookupEnv("DB_USER"); ok {
		DB_USER = strings.TrimSpace(dbUser)
	}

	if dbPassword, ok := os.LookupEnv("DB_PASSWORD"); ok {
		DB_PASSWORD = strings.TrimSpace(dbPassword)
	}

	if dbName, ok := os.LookupEnv("DB_NAME"); ok {
		DB_NAME = strings.TrimSpace(dbName)
	}

	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <init|up|down|status|mark_applied>", os.Args[0])
	}
}

func main() {
	pgconn := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(DB_URL),
		pgdriver.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		pgdriver.WithUser(DB_USER),
		pgdriver.WithPassword(DB_PASSWORD),
		pgdriver.WithDatabase(DB_NAME),
		pgdriver.WithInsecure(true),
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

	ctx := context.Background()
	cmd := os.Args[1]
	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if cmd == "init" {
		if err := migrator.Init(ctx); err != nil {
			panic(err)
		}
	} else if cmd == "up" {
		group, err := migrator.Migrate(context.Background())
		if err != nil {
			panic(err)
		}

		if group.ID == 0 {
			fmt.Printf("there are no new migrations to run\n")
			return
		}

		fmt.Printf("migrated to %s\n", group)
	} else if cmd == "down" {
		group, err := migrator.Rollback(ctx)
		if err != nil {
			panic(err)
		}

		if group.ID == 0 {
			fmt.Printf("there are no groups to roll back\n")
			return
		}

		fmt.Printf("rolled back %s\n", group)
	} else if cmd == "status" {
		ms, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("migrations: %s\n", ms)
		fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
		fmt.Printf("last migration group: %s\n", ms.LastGroup())
	} else if cmd == "mark_applied" {
		group, err := migrator.Migrate(ctx, migrate.WithNopMigration())
		if err != nil {
			panic(err)
		}

		if group.ID == 0 {
			fmt.Printf("there are no new migrations to mark as applied\n")
			return
		}

		fmt.Printf("marked as applied %s\n", group)
	}
}
