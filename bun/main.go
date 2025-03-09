package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	"my-dgii-api/cmd"
	"my-dgii-api/contributors"
	"my-dgii-api/database"
)

func main() {
        err := godotenv.Load()
        if err != nil {
                log.Fatal("Error loading .env file")
        }

        app := &cli.App{
                Name:  "my-dgii-api",
                Usage: "REST API with database migrations",
                Commands: []*cli.Command{
                        {
                                Name:  "serve",
                                Usage: "Start the HTTP server and schedule the DGII data import",
                                Action: func(c *cli.Context) error {
                                        router := contributors.InitModule()
                                        port := os.Getenv("PORT")
                                        address := fmt.Sprintf(":%s", port)

                                        fmt.Printf("Server listening on %s\n", address)

                                        // Schedule the weekly DGII data import task.
                                        scheduleDGIIImport()

                                        return http.ListenAndServe(address, router)
                                },
                        },
                        cmd.NewDBCommand(),
                },
        }

        if err := app.Run(os.Args); err != nil {
                log.Fatal(err)
        }
}

// scheduleDGIIImport schedules the weekly DGII data import task using gocron.
func scheduleDGIIImport() {
        db := database.ConnectDB() // Get the database connection.

        s := gocron.NewScheduler(time.UTC)

        // Schedule the task to run every Monday at 3:00 AM UTC.
        _, err := s.Cron("0 3 * * MON").Do(func() {
                log.Println("Executing DGII contributors import...")
                handler := contributors.NewContributorHandler(db) // Create a new ContributorHandler instance.
                err := handler.ImportContributorsFromDGII(db) // Call the import function.
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