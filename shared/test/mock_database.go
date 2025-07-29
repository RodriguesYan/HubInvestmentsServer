package test

import (
	"HubInvestments/shared/infra/database"
	"context"
	"database/sql"
	"reflect"

	"github.com/stretchr/testify/mock"
)

// MockDatabase implements the database.Database interface for testing
// This can be reused across all test files that need to mock database operations
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Query(query string, args ...interface{}) (database.Rows, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(database.Rows), arguments.Error(1)
}

func (m *MockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(database.Rows), arguments.Error(1)
}

func (m *MockDatabase) QueryRow(query string, args ...interface{}) database.Row {
	arguments := m.Called(query, args)
	return arguments.Get(0).(database.Row)
}

func (m *MockDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) database.Row {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(database.Row)
}

func (m *MockDatabase) Exec(query string, args ...interface{}) (database.Result, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(database.Result), arguments.Error(1)
}

func (m *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(database.Result), arguments.Error(1)
}

func (m *MockDatabase) Begin() (database.Transaction, error) {
	arguments := m.Called()
	return arguments.Get(0).(database.Transaction), arguments.Error(1)
}

func (m *MockDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (database.Transaction, error) {
	arguments := m.Called(ctx, opts)
	return arguments.Get(0).(database.Transaction), arguments.Error(1)
}

func (m *MockDatabase) Get(dest interface{}, query string, args ...interface{}) error {
	arguments := m.Called(dest, query, args)

	// If we're expecting a successful result and have mock data, populate the destination
	if arguments.Error(0) == nil && len(arguments) > 1 {
		if expectedData := arguments.Get(1); expectedData != nil {
			// Use reflection to copy the expected data to the destination
			destValue := reflect.ValueOf(dest)
			if destValue.Kind() == reflect.Ptr {
				destValue = destValue.Elem()
				expectedValue := reflect.ValueOf(expectedData)
				if destValue.Type() == expectedValue.Type() {
					destValue.Set(expectedValue)
				}
			}
		}
	}

	return arguments.Error(0)
}

func (m *MockDatabase) Select(dest interface{}, query string, args ...interface{}) error {
	arguments := m.Called(dest, query, args)

	// If we're expecting a successful result and have mock data, populate the destination
	if arguments.Error(0) == nil && len(arguments) > 1 {
		if expectedData := arguments.Get(1); expectedData != nil {
			// Use reflection to copy the expected data to the destination
			destValue := reflect.ValueOf(dest)
			if destValue.Kind() == reflect.Ptr {
				destValue = destValue.Elem()
				expectedValue := reflect.ValueOf(expectedData)
				if destValue.Type() == expectedValue.Type() {
					destValue.Set(expectedValue)
				}
			}
		}
	}

	return arguments.Error(0)
}

func (m *MockDatabase) Ping() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockDatabase) Close() error {
	arguments := m.Called()
	return arguments.Error(0)
}

// NewMockDatabase creates a new mock database instance
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}
