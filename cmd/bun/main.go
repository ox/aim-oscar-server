package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-bun/bun-starter-kit/bunapp"
	"github.com/go-bun/bun-starter-kit/cmd/bun/migrations"
	"github.com/go-bun/bun/migrate"

	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "bun",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "env",
				Value: "dev",
				Usage: "environment",
			},
		},
		Commands: []*cli.Command{
			newDBCommand(migrations.Migrations),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
func newDBCommand(migrations *migrate.Migrations) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "manage database migrations",
		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)
					return migrator.Init(ctx)
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)

					group, err := migrator.Migrate(ctx)
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Printf("there are no new migrations to run\n")
						return nil
					}

					fmt.Printf("migrated to %s\n", group)
					return nil
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)

					group, err := migrator.Rollback(ctx)
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Printf("there are no groups to roll back\n")
						return nil
					}

					fmt.Printf("rolled back %s\n", group)
					return nil
				},
			},
			{
				Name:  "lock",
				Usage: "lock migrations",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)
					return migrator.Lock(ctx)
				},
			},
			{
				Name:  "unlock",
				Usage: "unlock migrations",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)
					return migrator.Unlock(ctx)
				},
			},
			{
				Name:  "create_go",
				Usage: "create Go migration",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)

					name := strings.Join(c.Args().Slice(), "_")
					mf, err := migrator.CreateGoMigration(ctx, name)
					if err != nil {
						return err
					}
					fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)

					return nil
				},
			},
			{
				Name:  "create_sql",
				Usage: "create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)

					name := strings.Join(c.Args().Slice(), "_")
					files, err := migrator.CreateSQLMigrations(ctx, name)
					if err != nil {
						return err
					}

					for _, mf := range files {
						fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)

					ms, err := migrator.MigrationsWithStatus(ctx)
					if err != nil {
						return err
					}
					fmt.Printf("migrations: %s\n", ms)
					fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
					fmt.Printf("last migration group: %s\n", ms.LastGroup())

					return nil
				},
			},
			{
				Name:  "mark_applied",
				Usage: "mark migrations as applied without actually running them",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := migrate.NewMigrator(app.DB(), migrations)

					group, err := migrator.Migrate(ctx, migrate.WithNopMigration())
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Printf("there are no new migrations to mark as applied\n")
						return nil
					}

					fmt.Printf("marked as applied %s\n", group)
					return nil
				},
			},
		},
	}
}

func isServerClosed(err error) bool {
	return err.Error() == "http: Server closed"
}
