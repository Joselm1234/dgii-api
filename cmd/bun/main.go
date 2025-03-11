package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"

	"my-dgii-api/bunapp"
	"my-dgii-api/cmd/bun/migrations"
	"my-dgii-api/contributors"
	"my-dgii-api/httputil"
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
			apiCommand,
			newDBCommand(migrations.Migrations),
		},
	}
    if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
    var apiCommand = &cli.Command{
                Name:  "api",
                Usage: "start API server",
                Flags: []cli.Flag{
                        &cli.StringFlag{
                                Name:  "addr",
                                Value: ":8000",
                                Usage: "serve address",
                        },
                },
                Action: func(c *cli.Context) error {
                        // Start the application context
                        ctx, app, err := bunapp.StartCLI(c)
                        if err != nil {
                                return err
                        }
                        defer app.Stop()
        
                        // Set up the HTTP handler
                        var handler http.Handler
                        handler = app.Router()
                        handler = httputil.PanicHandler{Next: handler}
        
                        // Set up and start the HTTP server
                        srv := &http.Server{
                                Addr:         c.String("addr"),
                                ReadTimeout:  5 * time.Minute,
                                WriteTimeout: 10 * time.Minute,
                                IdleTimeout:  60 * time.Minute,
                                Handler:      handler,
                        }
        
                        // Start the main HTTP server
                        go func() {
                                if err := srv.ListenAndServe(); err != nil && !isServerClosed(err) {
                                        log.Printf("ListenAndServe failed: %s", err)
                                }
                        }()
                        
        
        
                        fmt.Printf("listening on %s\n", srv.Addr)
                        fmt.Println(bunapp.WaitExitSignal())
        
                        // Shut down the server
                        return srv.Shutdown(ctx)
                },
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

                            // Import initial DGII data after successful migration
					handler := contributors.NewContributorHandler(app)
					log.Println("Importing initial DGII data...")
					if err := handler.ImportContributorsFromDGII(); err != nil {
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
        // Function to check if the server is closed
func isServerClosed(err error) bool {
	return err.Error() == "http: Server closed"
}
        
// scheduleDGIIImport schedules the weekly DGII data import task using gocron.
func scheduleDGIIImport() {
        var app *bunapp.App

        s := gocron.NewScheduler(time.UTC)

        // Schedule the task to run every Monday at 3:00 AM UTC.
        _, err := s.Cron("0 3 * * MON").Do(func() {
                log.Println("Executing DGII contributors import...")
                handler := contributors.NewContributorHandler(app) // Create a new ContributorHandler instance.
                err := handler.ImportContributorsFromDGII() // Call the import function.
                if err != nil {
                        log.Printf("Error during import: %v", err)
                } else {
                        log.Println("Import completed.")
                }
        })

        if err != nil {
                log.Fatalf("Error scheduling task: %v", err)
        }

        s.StartAsync() // Start the scheduler asynchronously.
}