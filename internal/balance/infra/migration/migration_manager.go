package migration

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// BalanceMigrationManager handles database migrations for the balance module
type BalanceMigrationManager struct {
	migrate *migrate.Migrate
}

// NewBalanceMigrationManager creates a new migration manager for balance module
func NewBalanceMigrationManager(databaseURL string) (*BalanceMigrationManager, error) {
	// Connect to database to create the postgres driver instance
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Create postgres driver instance for migrations
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get absolute path to migrations directory
	migrationsPath, err := filepath.Abs("internal/balance/infra/migration/sql")
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &BalanceMigrationManager{migrate: m}, nil
}

// Up runs all pending migrations
func (bmm *BalanceMigrationManager) Up() error {
	err := bmm.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	return nil
}

// Down reverts the most recent migration
func (bmm *BalanceMigrationManager) Down() error {
	err := bmm.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration down: %w", err)
	}
	return nil
}

// Steps runs a specific number of migrations (positive for up, negative for down)
func (bmm *BalanceMigrationManager) Steps(n int) error {
	err := bmm.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d migration steps: %w", n, err)
	}
	return nil
}

// Force sets the migration version without running migrations (use with caution)
func (bmm *BalanceMigrationManager) Force(version int) error {
	err := bmm.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force migration to version %d: %w", version, err)
	}
	return nil
}

// Version returns the current migration version
func (bmm *BalanceMigrationManager) Version() (uint, bool, error) {
	version, dirty, err := bmm.migrate.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Close closes the migration instance
func (bmm *BalanceMigrationManager) Close() error {
	sourceErr, dbErr := bmm.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close migration database: %w", dbErr)
	}
	return nil
}
