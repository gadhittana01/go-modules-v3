package utils

import (
	"os"

	"github.com/joho/godotenv"
)

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
