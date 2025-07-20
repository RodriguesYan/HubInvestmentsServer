package database

import (
	"context"
	"database/sql"
)

// Database defines the interface for database operations
// This abstraction allows switching between different SQL packages (sqlx, sql, gorm, etc.)
// without changing repository implementations
type Database interface {
	// Query execution methods
	Query(query string, args ...interface{}) (Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(query string, args ...interface{}) Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row

	// Execute methods for INSERT, UPDATE, DELETE
	Exec(query string, args ...interface{}) (Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)

	// Transaction support
	Begin() (Transaction, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)

	// Convenience methods for common operations
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error

	// Connection management
	Ping() error
	Close() error
}

// Transaction represents a database transaction
type Transaction interface {
	// Query execution methods within transaction
	Query(query string, args ...interface{}) (Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(query string, args ...interface{}) Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row

	// Execute methods within transaction
	Exec(query string, args ...interface{}) (Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)

	// Convenience methods within transaction
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error

	// Transaction control
	Commit() error
	Rollback() error
}

// Rows represents the result of a query
type Rows interface {
	Close() error
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
	Columns() ([]string, error)
}

// Row represents the result of a single row query
type Row interface {
	Scan(dest ...interface{}) error
	Err() error
}

// Result represents the result of an exec operation
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}
