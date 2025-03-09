package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func ConnectDB() *bun.DB {
	dsn := os.Getenv("DATABASE_URL") // Example: "postgres://user:pass@localhost:5432/dbname"
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	db := bun.NewDB(sqlDB, pgdialect.New())
	return db
}
