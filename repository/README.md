# Repository Package

Base repository implementation following the book-go pattern.

## Pattern Overview

This follows the exact pattern from book-go:

```
DBTX (interface) → Queries (struct) → Repository (interface) → RepositoryImpl (struct)
```

## Files

- **db.go** - DBTX interface and Queries struct
- **base.go** - BaseRepository for common functionality
- **interfaces.go** - Base repository interfaces

## Usage in Services

### 1. Define Your Querier Interface

```go
// user/db/repository/querier.go
package repository

import (
	"context"
	"github.com/google/uuid"
)

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
}
```

### 2. Create Repository Interface

```go
// user/db/repository/repository.go
package repository

import (
	"github.com/jackc/pgx/v5"
	"github.com/yourusername/go-modules/utils"
)

type Repository interface {
	Querier
	
	WithTx(tx pgx.Tx) Querier
	GetDB() utils.PGXPool
}

type RepositoryImpl struct {
	db utils.PGXPool
	*Queries
}

func NewRepository(db utils.PGXPool) Repository {
	return &RepositoryImpl{
		db:      db,
		Queries: New(db),
	}
}

func (r *RepositoryImpl) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		db: tx,
	}
}

func (r *RepositoryImpl) GetDB() utils.PGXPool {
	return r.db
}
```

### 3. Implement Queries

```go
// user/db/repository/db.go
package repository

import (
	"context"
	
	base "github.com/yourusername/go-modules/repository"
)

// Use the base DBTX and Queries
type DBTX = base.DBTX

type Queries struct {
	db DBTX
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// Implement your query methods
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING *`
	var user User
	err := q.db.QueryRow(ctx, query, arg.Name, arg.Email, arg.Password).Scan(...)
	return user, err
}

func (q *Queries) FindUserByEmail(ctx context.Context, email string) (User, error) {
	query := `SELECT * FROM users WHERE email = $1`
	var user User
	err := q.db.QueryRow(ctx, query, email).Scan(...)
	return user, err
}
```

### 4. Use with Transactions

```go
import "github.com/yourusername/go-modules/utils"

// Execute with transaction
err := utils.ExecTxPool(ctx, repo.GetDB(), func(tx pgx.Tx) error {
	repoTx := repo.WithTx(tx)
	
	// All operations use the transaction
	user, err := repoTx.CreateUser(ctx, CreateUserParams{...})
	if err != nil {
		return err // Auto-rollback
	}
	
	// More operations...
	
	return nil // Auto-commit
})
```

## Complete Example

Here's a complete service repository structure:

```
user/
└── db/
    └── repository/
        ├── db.go           # DBTX type alias, Queries struct
        ├── querier.go      # Querier interface (all query methods)
        ├── repository.go   # Repository interface & implementation
        ├── models.go       # Data models
        └── user_queries.go # User query implementations
```

### db.go
```go
package repository

import (
	"context"
	base "github.com/yourusername/go-modules/repository"
)

type DBTX = base.DBTX

type Queries struct {
	db DBTX
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}
```

### querier.go
```go
package repository

import (
	"context"
	"github.com/google/uuid"
)

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	// ... more methods
}

var _ Querier = (*Queries)(nil)
```

### repository.go
```go
package repository

import (
	"github.com/jackc/pgx/v5"
	"github.com/yourusername/go-modules/utils"
)

type Repository interface {
	Querier
	WithTx(tx pgx.Tx) Querier
	GetDB() utils.PGXPool
}

type RepositoryImpl struct {
	db utils.PGXPool
	*Queries
}

func NewRepository(db utils.PGXPool) Repository {
	return &RepositoryImpl{db: db, Queries: New(db)}
}

func (r *RepositoryImpl) WithTx(tx pgx.Tx) Querier {
	return &Queries{db: tx}
}

func (r *RepositoryImpl) GetDB() utils.PGXPool {
	return r.db
}
```

### user_queries.go
```go
package repository

import "context"

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	query := `INSERT INTO users (name, email, password) 
	          VALUES ($1, $2, $3) 
	          RETURNING id, name, email, created_at`
	
	var user User
	err := q.db.QueryRow(ctx, query, arg.Name, arg.Email, arg.Password).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)
	return user, err
}
```

## Key Benefits

✅ **Exact book-go pattern** - Same structure and interfaces  
✅ **Transaction support** - WithTx pattern for safe transactions  
✅ **Type safety** - Strong typing with interfaces  
✅ **Testable** - Easy to mock for unit tests  
✅ **Consistent** - All services follow same pattern  
✅ **Reusable** - Shared DBTX and base implementations  

## Transaction Pattern

Always use transactions for multiple operations:

```go
err := utils.ExecTxPool(ctx, repo.GetDB(), func(tx pgx.Tx) error {
	repoTx := repo.WithTx(tx)
	
	// Operation 1
	user, err := repoTx.CreateUser(ctx, userParams)
	if err != nil {
		return err // Auto-rollback
	}
	
	// Operation 2
	_, err = repoTx.CreateProfile(ctx, ProfileParams{UserID: user.ID})
	if err != nil {
		return err // Auto-rollback
	}
	
	return nil // Auto-commit
})
```
