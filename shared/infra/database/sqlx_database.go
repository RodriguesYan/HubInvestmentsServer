package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// SQLXDatabase implements the Database interface using SQLX
type SQLXDatabase struct {
	db *sqlx.DB
}

// NewSQLXDatabase creates a new SQLX database implementation
func NewSQLXDatabase(db *sqlx.DB) Database {
	return &SQLXDatabase{db: db}
}

// Query executes a query and returns rows
func (s *SQLXDatabase) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLXRows{rows: rows}, nil
}

// QueryContext executes a query with context and returns rows
func (s *SQLXDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLXRows{rows: rows}, nil
}

// QueryRow executes a query and returns a single row
func (s *SQLXDatabase) QueryRow(query string, args ...interface{}) Row {
	row := s.db.QueryRow(query, args...)
	return &SQLXRow{row: row}
}

// QueryRowContext executes a query with context and returns a single row
func (s *SQLXDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	row := s.db.QueryRowContext(ctx, query, args...)
	return &SQLXRow{row: row}
}

// Exec executes a query without returning rows
func (s *SQLXDatabase) Exec(query string, args ...interface{}) (Result, error) {
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLXResult{result: result}, nil
}

// ExecContext executes a query with context without returning rows
func (s *SQLXDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLXResult{result: result}, nil
}

// Begin starts a transaction
func (s *SQLXDatabase) Begin() (Transaction, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	return &SQLXTransaction{tx: tx}, nil
}

// BeginTx starts a transaction with options
func (s *SQLXDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
	tx, err := s.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &SQLXTransaction{tx: tx}, nil
}

// Get executes a query and scans the result into dest
func (s *SQLXDatabase) Get(dest interface{}, query string, args ...interface{}) error {
	return s.db.Get(dest, query, args...)
}

// Select executes a query and scans the results into dest
func (s *SQLXDatabase) Select(dest interface{}, query string, args ...interface{}) error {
	return s.db.Select(dest, query, args...)
}

// Ping verifies the database connection
func (s *SQLXDatabase) Ping() error {
	return s.db.Ping()
}

// Close closes the database connection
func (s *SQLXDatabase) Close() error {
	return s.db.Close()
}

// SQLXTransaction implements the Transaction interface using SQLX
type SQLXTransaction struct {
	tx *sqlx.Tx
}

// Query executes a query within the transaction
func (t *SQLXTransaction) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLXRows{rows: rows}, nil
}

// QueryContext executes a query with context within the transaction
func (t *SQLXTransaction) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLXRows{rows: rows}, nil
}

// QueryRow executes a query and returns a single row within the transaction
func (t *SQLXTransaction) QueryRow(query string, args ...interface{}) Row {
	row := t.tx.QueryRow(query, args...)
	return &SQLXRow{row: row}
}

// QueryRowContext executes a query with context and returns a single row within the transaction
func (t *SQLXTransaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &SQLXRow{row: row}
}

// Exec executes a query within the transaction
func (t *SQLXTransaction) Exec(query string, args ...interface{}) (Result, error) {
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLXResult{result: result}, nil
}

// ExecContext executes a query with context within the transaction
func (t *SQLXTransaction) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLXResult{result: result}, nil
}

// Get executes a query and scans the result into dest within the transaction
func (t *SQLXTransaction) Get(dest interface{}, query string, args ...interface{}) error {
	return t.tx.Get(dest, query, args...)
}

// Select executes a query and scans the results into dest within the transaction
func (t *SQLXTransaction) Select(dest interface{}, query string, args ...interface{}) error {
	return t.tx.Select(dest, query, args...)
}

// Commit commits the transaction
func (t *SQLXTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *SQLXTransaction) Rollback() error {
	return t.tx.Rollback()
}

// SQLXRows implements the Rows interface using sql.Rows
type SQLXRows struct {
	rows *sql.Rows
}

// Close closes the rows
func (r *SQLXRows) Close() error {
	return r.rows.Close()
}

// Err returns the error, if any, that was encountered during iteration
func (r *SQLXRows) Err() error {
	return r.rows.Err()
}

// Next prepares the next result row for reading
func (r *SQLXRows) Next() bool {
	return r.rows.Next()
}

// Scan copies the columns in the current row into the values pointed at by dest
func (r *SQLXRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Columns returns the column names
func (r *SQLXRows) Columns() ([]string, error) {
	return r.rows.Columns()
}

// SQLXRow implements the Row interface using sql.Row
type SQLXRow struct {
	row *sql.Row
}

// Scan copies the columns from the matched row into the values pointed at by dest
func (r *SQLXRow) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// Err returns the error, if any, that was encountered during the query
func (r *SQLXRow) Err() error {
	return r.row.Err()
}

// SQLXResult implements the Result interface using sql.Result
type SQLXResult struct {
	result sql.Result
}

// LastInsertId returns the database's auto-generated ID after an INSERT into a table with an auto-increment primary key
func (r *SQLXResult) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}

// RowsAffected returns the number of rows affected by an update, insert, or delete
func (r *SQLXResult) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}
