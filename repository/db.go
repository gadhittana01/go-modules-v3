package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the interface for database operations
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// Queries is the base struct for all query operations
type Queries struct {
	db DBTX
}

// New creates a new Queries instance
func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// WithTx creates a new Queries instance with a transaction
func (q *Queries) WithTx(tx pgx.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}

// GetDB returns the database connection
func (q *Queries) GetDB() DBTX {
	return q.db
}
