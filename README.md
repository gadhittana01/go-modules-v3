# Go Modules - Shared Utilities

Common utilities and helpers for Cat and Dog microservices.

## Packages

### 📦 utils/
Common utility functions for all services.

- **database.go** - PostgreSQL connection pool and transaction helpers
- **redis.go** - Redis client initialization
- **token.go** - JWT generation and validation
- **crypto.go** - Password hashing with bcrypt
- **migration.go** - Database migration utilities
- **error.go** - Custom error handling
- **config.go** - Environment variable helpers

### 📦 repository/
Base repository interfaces and implementation.

- **base.go** - Base repository with common operations
- **interfaces.go** - Repository interfaces and contracts
- **README.md** - Detailed usage guide

## Installation

```bash
go get github.com/yourusername/go-modules
```

## Quick Start

### Database & Transactions

```go
import "github.com/yourusername/go-modules/utils"

pool, err := utils.ConnectDBPool(databaseURL)

// Execute transaction
err = utils.ExecTxPool(ctx, pool, func(tx pgx.Tx) error {
    // Your transaction logic
    return nil
})
```

### Repository Pattern

```go
import base "github.com/yourusername/go-modules/repository"

type Repository interface {
    YourQuerier
    base.Repository
    WithTx(tx pgx.Tx) YourQuerier
}

type RepositoryImpl struct {
    *base.BaseRepository
}

func NewRepository(pool interface{...}, schema string) Repository {
    return &RepositoryImpl{
        BaseRepository: base.NewBaseRepository(pool, schema),
    }
}
```

### JWT Token

```go
tokenClient := utils.NewToken("secret-key", 72)

// Generate
token, err := tokenClient.GenerateToken(utils.GenerateTokenReq{
    UserID: userID.String(),
})

// Validate
userID, err := tokenClient.ValidateToken(tokenString)
```

### Password Hashing

```go
hashedPassword, err := utils.HashPassword("mypassword")
isValid := utils.CheckPassword("mypassword", hashedPassword)
```

### Redis

```go
redisClient := utils.InitRedis(utils.RedisConfig{
    Host: "localhost",
    Port: "6379",
})
```

### Migration

```go
err := utils.RunMigration(databaseURL, "schema_name", "db/migration")
```

### Error Handling

```go
// Custom errors with status codes
err := utils.NewCustomError("not found", 404)
err := utils.NewCustomErrorWithTrace(err, "failed to get user", 400)

// Panic helpers
utils.PanicIfError(err)
utils.PanicIfAppError(err, "operation failed", 500)
```

## Features

✅ **Database** - PGX connection pool and transaction management  
✅ **Repository** - Base repository pattern with transaction support  
✅ **Redis** - Redis client initialization  
✅ **JWT** - Token generation and validation  
✅ **Crypto** - Password hashing with bcrypt  
✅ **Migration** - Database migration runner  
✅ **Error** - Custom error handling with status codes  
✅ **Config** - Environment variable helpers  

## Structure

```
go-modules/
├── go.mod
├── README.md
├── utils/
│   ├── database.go      # PGX pool & transactions
│   ├── redis.go         # Redis client
│   ├── token.go         # JWT utilities
│   ├── crypto.go        # Password hashing
│   ├── migration.go     # DB migrations
│   ├── error.go         # Error handling
│   └── config.go        # Config helpers
└── repository/
    ├── base.go          # Base repository
    ├── interfaces.go    # Repository interfaces
    └── README.md        # Usage guide
```

## License

MIT
