package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type JobState string

const (
	StatusPending   JobState = "PENDING"
	StatusRunning   JobState = "RUNNING"
	StatusSuccess   JobState = "SUCCESS"
	StatusFailed    JobState = "FAILED"
	StatusDead      JobState = "DEAD"
)

type JobRecord struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Payload    []byte    `db:"payload"`
	Status     JobState  `db:"status"`
	RetryCount int       `db:"retry_count"`
	MaxRetries int       `db:"max_retries"`
	ErrorLog   string    `db:"error_log"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo initializes the sql.DB connection pool
func NewPostgresRepo(dataSourceName string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database pool: %w", err)
	}

	// Production connection pool tuning
	db.SetMaxOpenConns(25)                 // Prevent exhausting DB descriptors
	db.SetMaxIdleConns(10)                 // Keep hot connections ready
	db.SetConnMaxLifetime(5 * time.Minute) // Cycle old connections out

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	return &PostgresRepo{db: db}, nil
}

// CreateJob inserts the initial job entry as PENDING
func (r *PostgresRepo) CreateJob(ctx context.Context, id, name string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal state payload: %w", err)
	}

	query := `
		INSERT INTO jobs (id, name, payload, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'PENDING', NOW(), NOW())
	`
	_, err = r.db.ExecContext(ctx, query, id, name, payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to insert job record: %w", err)
	}
	return nil
}

// UpdateJobStatus modifies the state of a specific job
func (r *PostgresRepo) UpdateJobStatus(ctx context.Context, id string, status JobState, errMsg string) error {
	var query string
	var err error

	if errMsg != "" {
		query = `
			UPDATE jobs
			SET status = $1, error_log = $2, updated_at = NOW()
			WHERE id = $3
		`
		_, err = r.db.ExecContext(ctx, query, status, errMsg, id)
	} else {
		query = `
			UPDATE jobs
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`
		_, err = r.db.ExecContext(ctx, query, status, id)
	}

	if err != nil {
		return fmt.Errorf("failed to update state for job %s: %w", id, err)
	}
	return nil
}

// GetJob reads a single job record directly from the read index
func (r *PostgresRepo) GetJob(ctx context.Context, id string) (*JobRecord, error) {
	query := `
		SELECT id, name, payload, status, retry_count, max_retries, COALESCE(error_log, ''), created_at, updated_at 
		FROM jobs
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var j JobRecord
	err := row.Scan(&j.ID, &j.Name, &j.Payload, &j.Status, &j.RetryCount, &j.MaxRetries, &j.ErrorLog, &j.CreatedAt, &j.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	return &j, nil
}

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// IncrementRetry increments the counter and saves the latest execution error
func (r *PostgresRepo) IncrementRetry(ctx context.Context, id string, errMsg string) (int, int, error) {
	query := `
		UPDATE jobs
		SET retry_count = retry_count + 1, error_log = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING retry_count, max_retries
	`
	var currentRetry, maxRetries int
	err := r.db.QueryRowContext(ctx, query, errMsg, id).Scan(&currentRetry, &maxRetries)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to increment retry parameters: %w", err)
	}
	return currentRetry, maxRetries, nil
}
