package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type PGXPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Close()
}

// ConnectDBPool creates a new database connection pool with retry logic
func ConnectDBPool(databaseURL string) (PGXPool, error) {
	var dbPool *pgxpool.Pool
	var err error

	// Retry logic for database connection
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		dbPool, err = pgxpool.New(context.Background(), databaseURL)
		if err == nil {
			// Test the connection
			err = dbPool.Ping(context.Background())
			if err == nil {
				log.Println("Successfully connected to database")
				return dbPool, nil
			}
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * 2)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// ConnectDB creates a sql.DB connection for migrations
func ConnectDB(databaseURL string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// Retry logic for database connection
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", databaseURL)
		if err == nil {
			// Test the connection
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to database (sql.DB)")
				return db, nil
			}
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * 2)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// ExecTxPool executes a function within a database transaction
func ExecTxPool(ctx context.Context, pool PGXPool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
