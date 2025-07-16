package persistence

import (
	"errors"
	"testing"

	"HubInvestments/position/infra/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLXPositionRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewSQLXPositionRepository(sqlxDB)

	assert.NotNil(t, repo)

	// Type assertion to check if we have the correct implementation
	sqlxRepo, ok := repo.(*SQLXPositionRepository)
	require.True(t, ok)

	assert.Equal(t, sqlxDB, sqlxRepo.db)
	assert.NotNil(t, sqlxRepo.mapper)
}

func TestSQLXPositionRepository_GetPositionsByUserId(t *testing.T) {
	// Common setup
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewSQLXPositionRepository(sqlxDB)

	testUserID := "user-123"

	// This regex is designed to be flexible with respect to whitespace in the SQL query,
	// making the test more robust against formatting changes.
	queryRegex := `SELECT\s+i.symbol,\s+p.average_price,\s+p.quantity,\s+i.category,\s+i.last_price\s+FROM\s+positions\s+p\s+JOIN\s+instruments\s+i\s+ON\s+p.instrument_id\s+=\s+i.id\s+WHERE\s+p.user_id\s+=\s+\$1`

	t.Run("success - returns positions", func(t *testing.T) {
		// Arrange
		expectedDTOs := []dto.AssetDTO{
			{Symbol: "AAPL", AveragePrice: 150.0, Quantity: 10, Category: 1, LastPrice: 170.0},
			{Symbol: "GOOG", AveragePrice: 2800.0, Quantity: 2, Category: 1, LastPrice: 2900.0},
		}

		rows := sqlmock.NewRows([]string{"symbol", "average_price", "quantity", "category", "last_price"}).
			AddRow(expectedDTOs[0].Symbol, expectedDTOs[0].AveragePrice, expectedDTOs[0].Quantity, expectedDTOs[0].Category, expectedDTOs[0].LastPrice).
			AddRow(expectedDTOs[1].Symbol, expectedDTOs[1].AveragePrice, expectedDTOs[1].Quantity, expectedDTOs[1].Category, expectedDTOs[1].LastPrice)

		mock.ExpectQuery(queryRegex).WithArgs(testUserID).WillReturnRows(rows)

		// Act
		positions, err := repo.GetPositionsByUserId(testUserID)

		// Assert
		require.NoError(t, err)

		// We use the repo's own mapper to generate the expected result.
		// This ensures we're testing the repository's behavior, including its use of the mapper.
		expectedModels := repo.(*SQLXPositionRepository).mapper.ToDomainSlice(expectedDTOs)
		assert.Equal(t, expectedModels, positions)
		assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations were not met")
	})

	t.Run("success - no positions found", func(t *testing.T) {
		// Arrange
		rows := sqlmock.NewRows([]string{"symbol", "average_price", "quantity", "category", "last_price"})
		mock.ExpectQuery(queryRegex).WithArgs(testUserID).WillReturnRows(rows)

		// Act
		positions, err := repo.GetPositionsByUserId(testUserID)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, positions)
		assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations were not met")
	})

	t.Run("error - database query fails", func(t *testing.T) {
		// Arrange
		dbError := errors.New("database connection lost")
		mock.ExpectQuery(queryRegex).WithArgs(testUserID).WillReturnError(dbError)

		// Act
		positions, err := repo.GetPositionsByUserId(testUserID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, positions)
		assert.ErrorIs(t, err, dbError)
		assert.Contains(t, err.Error(), "failed to get positions for user user-123")
		assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations were not met")
	})
}
