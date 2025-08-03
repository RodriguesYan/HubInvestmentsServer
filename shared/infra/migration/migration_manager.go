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

// MigrationManager handles database migrations for the entire HubInvestments project
type MigrationManager struct {
	migrate *migrate.Migrate
}

// NewMigrationManager creates a new migration manager for the entire project
func NewMigrationManager(databaseURL string) (*MigrationManager, error) {
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
	migrationsPath, err := filepath.Abs("shared/infra/migration/sql")
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

	return &MigrationManager{migrate: m}, nil
}

// Up runs all pending migrations
func (mm *MigrationManager) Up() error {
	err := mm.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	return nil
}

// Down reverts the most recent migration
func (mm *MigrationManager) Down() error {
	err := mm.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration down: %w", err)
	}
	return nil
}

// Steps runs a specific number of migrations (positive for up, negative for down)
func (mm *MigrationManager) Steps(n int) error {
	err := mm.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d migration steps: %w", n, err)
	}
	return nil
}

// Force sets the migration version without running migrations (use with caution)
func (mm *MigrationManager) Force(version int) error {
	err := mm.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force migration to version %d: %w", version, err)
	}
	return nil
}

// Version returns the current migration version
func (mm *MigrationManager) Version() (uint, bool, error) {
	version, dirty, err := mm.migrate.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Close closes the migration instance
func (mm *MigrationManager) Close() error {
	sourceErr, dbErr := mm.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close migration database: %w", dbErr)
	}
	return nil
}
