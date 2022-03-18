package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Load env
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	// Init database
	db, err := sql.Open("sqlite3", os.Getenv("DATABASE_DSN"))
	if err != nil {
		zap.S().Fatalf("DB schema sync error: %s", err)
	}
	defer db.Close()

	statement, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS rates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source TEXT,
			currency_from TEXT,
			currency_to TEXT,
			rate REAL,
			updated_at TIMESTAMP)
	`)
	if _, err := statement.Exec(); err != nil {
		zap.S().Fatalf("DB schema sync error: %s", err)
	}

	fmt.Println("DB schema sync ok")
	os.Exit(0)
}
