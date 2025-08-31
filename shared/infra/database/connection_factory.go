package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Driver   string
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

// ConnectionFactory manages database connections and provides the abstracted Database interface
type ConnectionFactory struct {
	config ConnectionConfig
}

func NewConnectionFactory(config ConnectionConfig) *ConnectionFactory {
	return &ConnectionFactory{config: config}
}

// CreateConnection creates a new database connection using SQLX
// This method encapsulates the SQLX package usage
// To switch to a different SQL package, only this method needs to be modified
func (cf *ConnectionFactory) CreateConnection() (Database, error) {
	switch cf.config.Driver {
	case "postgres":
		return cf.createPostgreSQLConnection()
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cf.config.Driver)
	}
}

// createPostgreSQLConnection creates a PostgreSQL connection using SQLX
func (cf *ConnectionFactory) createPostgreSQLConnection() (Database, error) {
	dsn := cf.buildPostgreSQLDSN()

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return NewSQLXDatabase(db), nil
}

// buildPostgreSQLDSN builds the PostgreSQL data source name
func (cf *ConnectionFactory) buildPostgreSQLDSN() string {
	dsn := fmt.Sprintf("user=%s dbname=%s sslmode=%s",
		cf.config.Username,
		cf.config.Database,
		cf.config.SSLMode,
	)

	if cf.config.Password != "" {
		dsn += fmt.Sprintf(" password=%s", cf.config.Password)
	}

	if cf.config.Host != "" {
		dsn += fmt.Sprintf(" host=%s", cf.config.Host)
	}

	if cf.config.Port != "" {
		dsn += fmt.Sprintf(" port=%s", cf.config.Port)
	}

	return dsn
}

// DefaultConfig returns a default configuration for local development
// This matches the current configuration used in the codebase
func DefaultConfig() ConnectionConfig {
	return ConnectionConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Database: "yanrodrigues",
		Username: "yanrodrigues",
		Password: "",
		SSLMode:  "disable",
	}
}

// CreateDatabaseConnection is a convenience function that creates a database connection
// with the default configuration. This can be used as a drop-in replacement for
// the current sqlx.Connect calls throughout the codebase
func CreateDatabaseConnection() (Database, error) {
	factory := NewConnectionFactory(DefaultConfig())
	return factory.CreateConnection()
}

// CreateDatabaseConnectionWithConfig creates a database connection with custom configuration
func CreateDatabaseConnectionWithConfig(config ConnectionConfig) (Database, error) {
	factory := NewConnectionFactory(config)
	return factory.CreateConnection()
}
