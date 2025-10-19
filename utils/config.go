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

// Config holds application configuration
type Config struct {
	DBConnString string
	RedisHost    string
	RedisPort    string
	Port         string
	MigrationURL string
	DBName       string
	JWTSecret    string
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

// CheckAndSetConfig loads configuration from environment variables
func CheckAndSetConfig(configPath, configName string) *Config {
	// Load environment variables from .env file
	LoadEnv(configPath + "/" + configName + ".env")

	return &Config{
		DBConnString: GetEnv("DB_CONN_STRING", "postgres://user:password@localhost:5432/dbname?sslmode=disable"),
		RedisHost:    GetEnv("REDIS_HOST", "localhost"),
		RedisPort:    GetEnv("REDIS_PORT", "6379"),
		Port:         GetEnv("PORT", "8000"),
		MigrationURL: "file://db/migration",
		DBName:       "postgres",
		JWTSecret:    GetEnv("JWT_SECRET", "your_jwt_secret_key_here_change_in_production"),
	}
}
