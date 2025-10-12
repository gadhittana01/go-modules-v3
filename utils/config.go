package utils

import (
	"os"

	"github.com/joho/godotenv"
)

// BaseConfig holds base configuration for migrations
type BaseConfig struct {
	MigrationURL string
	DBName       string
}

// LoadEnv loads environment variables from .env file
func LoadEnv(paths ...string) {
	if len(paths) > 0 {
		godotenv.Load(paths...)
	} else {
		godotenv.Load()
	}
}

// GetEnv gets an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// FormatTableName formats a table name with schema prefix
func FormatTableName(schema, table string) string {
	if schema != "" {
		return schema + "." + table
	}
	return table
}
