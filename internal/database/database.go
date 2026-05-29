package database

import (
	"database/sql"
	"os"
	"log"
	_ "github.com/lib/pq"
)

func Connect() *sql.DB {
	// 1. Explicitly check for the Docker Compose injected configuration variable
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		// Fall back to local execution ONLY if the environment variable is empty
		dbURL = "postgres://orchestrator_user:orchestrator_password@localhost:5432/job_orchestrator?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Fatal configuration error: %v", err)
	}

	return db
}
