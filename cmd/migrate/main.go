package main

import (
	"flag"
	"fmt"
	"log"

	"HubInvestments/internal/balance/infra/migration"
	"HubInvestments/shared/infra/database"
)

const (
	defaultDatabaseURL = "postgres://yanrodrigues@localhost/yanrodrigues?sslmode=disable"
)

func main() {
	var (
		command     = flag.String("command", "up", "Migration command: up, down, steps, force, version")
		module      = flag.String("module", "balance", "Module to migrate: balance")
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

	// Validate module
	if *module != "balance" {
		log.Fatalf("Unsupported module: %s. Supported modules: balance", *module)
	}

	// Run migration based on module
	switch *module {
	case "balance":
		err := runBalanceMigration(*command, *steps, *version, dbURL)
		if err != nil {
			log.Fatalf("Balance migration failed: %v", err)
		}
	default:
		log.Fatalf("Unknown module: %s", *module)
	}
}

func runBalanceMigration(command string, steps, version int, databaseURL string) error {
	mgr, err := migration.NewBalanceMigrationManager(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create balance migration manager: %w", err)
	}
	defer mgr.Close()

	switch command {
	case "up":
		fmt.Println("Running balance migrations up...")
		err = mgr.Up()
		if err != nil {
			return fmt.Errorf("failed to run migrations up: %w", err)
		}
		fmt.Println("✅ Balance migrations completed successfully")

	case "down":
		fmt.Println("Running balance migration down...")
		err = mgr.Down()
		if err != nil {
			return fmt.Errorf("failed to run migration down: %w", err)
		}
		fmt.Println("✅ Balance migration rolled back successfully")

	case "steps":
		fmt.Printf("Running %d migration steps...\n", steps)
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
