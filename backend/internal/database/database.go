package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Register the pgx stdlib database/sql driver.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/cipherkeep/backend/internal/domain"
)

// Connect opens a *sql.DB pool using the pgx stdlib driver and verifies it.
//
// The initial ping is retried with a fixed backoff for up to ~30 seconds so the
// service tolerates the database (or container networking) not being ready yet at
// startup, e.g. when started together with PostgreSQL under docker-compose.
func Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	const (
		maxAttempts = 15
		retryDelay  = 2 * time.Second
	)
	var pingErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr = db.PingContext(pingCtx)
		cancel()
		if pingErr == nil {
			return db, nil
		}
		if ctx.Err() != nil {
			break
		}
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
			case <-time.After(retryDelay):
			}
		}
	}

	_ = db.Close()
	return nil, fmt.Errorf("ping database after %d attempts: %w", maxAttempts, pingErr)
}

// WithTx runs fn inside a database transaction, committing on success and
// rolling back on error. It exposes a domain.Querier (the *sql.Tx) so
// repositories can participate in the unit of work.
func WithTx(ctx context.Context, db *sql.DB, fn func(q domain.Querier) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		// Rollback is a no-op if the tx was already committed.
		_ = tx.Rollback()
	}()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
