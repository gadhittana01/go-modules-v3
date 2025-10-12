package repository

import (
	"github.com/gadhittana01/go-modules-v3/utils"
)

// BaseRepository provides common repository functionality
// This is the base struct that service repositories should embed
type BaseRepository struct {
	db utils.PGXPool
	*Queries
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db utils.PGXPool) *BaseRepository {
	return &BaseRepository{
		db:      db,
		Queries: New(db),
	}
}

// GetDB returns the database pool
func (r *BaseRepository) GetDB() utils.PGXPool {
	return r.db
}
