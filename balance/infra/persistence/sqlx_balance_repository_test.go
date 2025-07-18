package persistence

import (
	"errors"
	"testing"

	domain "HubInvestments/balance/domain/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSqlxBalanceRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewSqlxBalanceRepository(sqlxDB)

	assert.NotNil(t, repo)

	// Type assertion to check if we have the correct implementation
	sqlxRepo, ok := repo.(*SQLXBalanceRepository)
	require.True(t, ok)

	assert.Equal(t, sqlxDB, sqlxRepo.db)
}

func TestSQLXBalanceRepository_GetBalance(t *testing.T) {
	// Common setup
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewSqlxBalanceRepository(sqlxDB)

	testUserID := "user-123"

	// This regex is designed to be flexible with respect to whitespace in the SQL query,
	// making the test more robust against formatting changes.
	queryRegex := `SELECT\s+available_balance\s+FROM\s+balances\s+WHERE\s+user_id\s+=\s+\$1`

	t.Run("success - returns balance", func(t *testing.T) {
		// Arrange
		expectedBalance := float32(1500.50)

		rows := sqlmock.NewRows([]string{"available_balance"}).
			AddRow(expectedBalance)

		mock.ExpectQuery(queryRegex).WithArgs(testUserID).WillReturnRows(rows)

		// Act
		balance, err := repo.GetBalance(testUserID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedBalance, balance.AvailableBalance)
		assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations were not met")
	})

	t.Run("success - no balance found", func(t *testing.T) {
		// Arrange
		rows := sqlmock.NewRows([]string{"available_balance"})
		mock.ExpectQuery(queryRegex).WithArgs(testUserID).WillReturnRows(rows)

		// Act
		balance, err := repo.GetBalance(testUserID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, domain.BalanceModel{}, balance)
		assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations were not met")
	})

	t.Run("error - database query fails", func(t *testing.T) {
		// Arrange
		dbError := errors.New("database connection lost")
		mock.ExpectQuery(queryRegex).WithArgs(testUserID).WillReturnError(dbError)

		// Act
		balance, err := repo.GetBalance(testUserID)

		// Assert
		require.Error(t, err)
		assert.Equal(t, domain.BalanceModel{}, balance)
		assert.Contains(t, err.Error(), "database connection lost")
		assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations were not met")
	})

	t.Run("error - row scan fails", func(t *testing.T) {
		// Arrange
		rows := sqlmock.NewRows([]string{"available_balance"}).
			AddRow("invalid_float_value") // This will cause a scan error

		mock.ExpectQuery(queryRegex).WithArgs(testUserID).WillReturnRows(rows)

		// Act
		balance, err := repo.GetBalance(testUserID)

		// Assert
		require.Error(t, err)
		assert.Equal(t, domain.BalanceModel{}, balance)
		assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations were not met")
	})

	t.Run("success - multiple rows returns last one", func(t *testing.T) {
		// Arrange
		firstBalance := float32(1000.0)
		expectedBalance := float32(2000.0)

		rows := sqlmock.NewRows([]string{"available_balance"}).
			AddRow(firstBalance).
			AddRow(expectedBalance)

		mock.ExpectQuery(queryRegex).WithArgs(testUserID).WillReturnRows(rows)

		// Act
		balance, err := repo.GetBalance(testUserID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedBalance, balance.AvailableBalance)
		assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations were not met")
	})
}
