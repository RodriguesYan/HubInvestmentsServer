package main

import (
	"flag"
	"fmt"
	"log"

	"HubInvestments/shared/infra/database"
	"HubInvestments/shared/infra/migration"
)

const (
	defaultDatabaseURL = "postgres://yanrodrigues@localhost/yanrodrigues?sslmode=disable"
)

func main() {
	var (
		command     = flag.String("command", "up", "Migration command: up, down, steps, force, version")
		steps       = flag.Int("steps", 1, "Number of steps for 'steps' command")
		version     = flag.Int("version", 0, "Version for 'force' command")
		databaseURL = flag.String("db", "", "Database URL (defaults to local development)")
	)
	flag.Parse()

	// Use default database URL if not provided
	dbURL := *databaseURL
	if dbURL == "" {
		dbURL = defaultDatabaseURL
	}

	// Run migrations
	err := runMigration(*command, *steps, *version, dbURL)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

func runMigration(command string, steps, version int, databaseURL string) error {
	mgr, err := migration.NewMigrationManager(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}
	defer mgr.Close()

	switch command {
	case "up":
		fmt.Println("Running all pending migrations...")
		err = mgr.Up()
		if err != nil {
			return fmt.Errorf("failed to run migrations up: %w", err)
		}
		fmt.Println("✅ All migrations completed successfully")

	case "down":
		fmt.Println("Rolling back the most recent migration...")
		err = mgr.Down()
		if err != nil {
			return fmt.Errorf("failed to run migration down: %w", err)
		}
		fmt.Println("✅ Migration rolled back successfully")

	case "steps":
		if steps > 0 {
			fmt.Printf("Running %d migration steps forward...\n", steps)
		} else {
			fmt.Printf("Running %d migration steps backward...\n", -steps)
		}
		err = mgr.Steps(steps)
		if err != nil {
			return fmt.Errorf("failed to run migration steps: %w", err)
		}
		fmt.Printf("✅ Completed %d migration steps successfully\n", steps)

	case "force":
		fmt.Printf("Forcing migration to version %d...\n", version)
		err = mgr.Force(version)
		if err != nil {
			return fmt.Errorf("failed to force migration: %w", err)
		}
		fmt.Printf("✅ Forced migration to version %d successfully\n", version)

	case "version":
		v, dirty, err := mgr.Version()
		if err != nil {
			return fmt.Errorf("failed to get migration version: %w", err)
		}
		fmt.Printf("Current migration version: %d", v)
		if dirty {
			fmt.Printf(" (dirty)")
		}
		fmt.Println()

	default:
		return fmt.Errorf("unknown command: %s. Available commands: up, down, steps, force, version", command)
	}

	return nil
}

func buildDatabaseURL() string {
	config := database.DefaultConfig()
	return fmt.Sprintf("postgres://%s@%s/%s?sslmode=%s",
		config.Username,
		config.Host,
		config.Database,
		config.SSLMode,
	)
}
