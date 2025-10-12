package utils

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigration runs database migrations
func RunMigration(databaseURL, schema, migrationPath string) error {
	migrationURL := databaseURL
	if schema != "" {
		migrationURL = fmt.Sprintf("%s?search_path=%s", databaseURL, schema)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		migrationURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply")
	} else {
		log.Println("Migrations applied successfully")
	}

	return nil
}
