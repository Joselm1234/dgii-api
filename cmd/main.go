package cmd

import (
	"context"
	"fmt"
	"log"
	"my-dgii-api/contributors"
	"my-dgii-api/database"

	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"
)

func NewDBCommand() *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "manage database migrations",
		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					db := database.ConnectDB()
					migrator := migrate.NewMigrator(db, Migrations)
					return migrator.Init(ctx)
				},
			},
			{
				Name:  "migrate",
				Usage: "apply new migrations",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					db := database.ConnectDB()
					migrator := migrate.NewMigrator(db, Migrations)

					group, err := migrator.Migrate(ctx)
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Println("No new migrations to apply.")
						return nil
					}

					fmt.Printf("Migrated to %s\n", group)

					// Import initial DGII data after successful migration
					handler := contributors.NewContributorHandler(db)
					log.Println("Importing initial DGII data...")
					if err := handler.ImportContributorsFromDGII(db); err != nil {
						log.Printf("Error during initial import: %v", err)
						// You might want to handle this error more gracefully,
						// depending on your application's requirements.
						// For example, you could return the error here,
						// but that would prevent the program from continuing.
						// For now, we'll just log the error and continue.
					} else {
						log.Println("Initial DGII import completed.")
					}

					return nil
				
				},

			},
			{
				Name:  "rollback",
				Usage: "rollback last migration group",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					db := database.ConnectDB()
					migrator := migrate.NewMigrator(db, Migrations)

					group, err := migrator.Rollback(ctx)
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Println("No migrations to roll back.")
						return nil
					}

					fmt.Printf("Rolled back %s\n", group)
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "check migration status",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					db := database.ConnectDB()
					migrator := migrate.NewMigrator(db, Migrations)

					ms, err := migrator.MigrationsWithStatus(ctx)
					if err != nil {
						return err
					}
					fmt.Println("Migrations:", ms)
					fmt.Println("Unapplied migrations:", ms.Unapplied())
					fmt.Println("Last migration group:", ms.LastGroup())

					return nil
				},
			},
		},
	}
}
