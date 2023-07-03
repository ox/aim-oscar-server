package main

import (
	"aim-oscar/cmd/migrate/migrations"
	"aim-oscar/config"
	"aim-oscar/db"
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/uptrace/bun/migrate"
)

func usage() {
	flag.Usage()
	log.Fatalf("Usage: migrate --config <config path> <init|up|down|status|mark_applied>\n")
}

func main() {
	configPath := flag.String("config", "", "Path to app config")
	flag.Parse()

	if configPath == nil || *configPath == "" {
		usage()
	}

	conf, err := config.FromFile(*configPath)
	if err != nil {
		log.Fatalf("could not parse config: %s", err)
	}

	db, err := db.Connect(&conf.DBConfig)
	if err != nil {
		log.Fatalf("could not connect to DB: %s", err)
	}

	ctx := context.Background()
	cmd := flag.Arg(0)
	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if cmd == "" {
		log.Println("Missing command")
		usage()
	}

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
