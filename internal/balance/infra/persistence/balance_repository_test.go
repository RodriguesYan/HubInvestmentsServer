package persistence

import (
	domain "HubInvestments/internal/balance/domain/model"
	repository "HubInvestments/internal/balance/domain/repository"
	"HubInvestments/shared/test"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewBalanceRepository(t *testing.T) {
	// Arrange
	mockDB := test.NewMockDatabase()

	// Act
	repo := NewBalanceRepository(mockDB)

	// Assert
	assert.NotNil(t, repo)
	assert.Implements(t, (*repository.IBalanceRepository)(nil), repo)
}

func TestBalanceRepository_GetBalance_Success(t *testing.T) {
	// Arrange
	mockDB := test.NewMockDatabase()
	repo := NewBalanceRepository(mockDB)

	userId := "user123"
	expectedBalance := float32(15000.50)
	expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`

	// Configure mock to return expected balance - pass expected data as second return value
	mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
		Return(nil, expectedBalance)

	// Act
	result, err := repo.GetBalance(userId)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, result.AvailableBalance)
	mockDB.AssertExpectations(t)
}

func TestBalanceRepository_GetBalance_NoRows(t *testing.T) {
	// Arrange
	mockDB := test.NewMockDatabase()
	repo := NewBalanceRepository(mockDB)

	userId := "nonexistent-user"
	expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`

	// Configure mock to return sql.ErrNoRows
	mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
		Return(sql.ErrNoRows)

	// Act
	result, err := repo.GetBalance(userId)

	// Assert
	assert.NoError(t, err)                               // Should not return error for no rows
	assert.Equal(t, float32(0), result.AvailableBalance) // Should return zero balance
	mockDB.AssertExpectations(t)
}

func TestBalanceRepository_GetBalance_DatabaseError(t *testing.T) {
	// Arrange
	mockDB := test.NewMockDatabase()
	repo := NewBalanceRepository(mockDB)

	userId := "user123"
	expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`
	dbError := errors.New("database connection failed")

	// Configure mock to return database error
	mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
		Return(dbError)

	// Act
	result, err := repo.GetBalance(userId)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get balance for user user123")
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Equal(t, domain.BalanceModel{}, result) // Should return empty model
	mockDB.AssertExpectations(t)
}

func TestBalanceRepository_GetBalance_EmptyUserId(t *testing.T) {
	// Arrange
	mockDB := test.NewMockDatabase()
	repo := NewBalanceRepository(mockDB)

	userId := ""
	expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`

	// Configure mock to return no rows for empty user ID
	mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
		Return(sql.ErrNoRows)

	// Act
	result, err := repo.GetBalance(userId)

	// Assert
	assert.NoError(t, err) // Should handle gracefully
	assert.Equal(t, float32(0), result.AvailableBalance)
	mockDB.AssertExpectations(t)
}

func TestBalanceRepository_GetBalance_ZeroBalance(t *testing.T) {
	// Arrange
	mockDB := test.NewMockDatabase()
	repo := NewBalanceRepository(mockDB)

	userId := "user-with-zero-balance"
	expectedBalance := float32(0.0)
	expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`

	// Configure mock to return zero balance
	mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
		Return(nil, expectedBalance)

	// Act
	result, err := repo.GetBalance(userId)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, float32(0.0), result.AvailableBalance)
	mockDB.AssertExpectations(t)
}

func TestBalanceRepository_GetBalance_NegativeBalance(t *testing.T) {
	// Arrange
	mockDB := test.NewMockDatabase()
	repo := NewBalanceRepository(mockDB)

	userId := "user-with-negative-balance"
	expectedBalance := float32(-500.25)
	expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`

	// Configure mock to return negative balance
	mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
		Return(nil, expectedBalance)

	// Act
	result, err := repo.GetBalance(userId)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, float32(-500.25), result.AvailableBalance)
	mockDB.AssertExpectations(t)
}

func TestBalanceRepository_GetBalance_LargeBalance(t *testing.T) {
	// Arrange
	mockDB := test.NewMockDatabase()
	repo := NewBalanceRepository(mockDB)

	userId := "wealthy-user"
	expectedBalance := float32(9999999.99)
	expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`

	// Configure mock to return large balance
	mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
		Return(nil, expectedBalance)

	// Act
	result, err := repo.GetBalance(userId)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, result.AvailableBalance)
	mockDB.AssertExpectations(t)
}

func TestBalanceRepository_GetBalance_SpecialCharacterUserId(t *testing.T) {
	// Test with user IDs containing special characters
	testCases := []struct {
		name   string
		userId string
	}{
		{"user ID with spaces", "user 123"},
		{"user ID with special chars", "user@#$%^&*()"},
		{"user ID with unicode", "用户123"},
		{"user ID with quotes", "user'123\"test"},
		{"UUID format", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockDB := test.NewMockDatabase()
			repo := NewBalanceRepository(mockDB)

			expectedBalance := float32(1234.56)
			expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`

			// Configure mock
			mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{tc.userId}).
				Return(nil, expectedBalance)

			// Act
			result, err := repo.GetBalance(tc.userId)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedBalance, result.AvailableBalance)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestBalanceRepository_GetBalance_ErrorHandling(t *testing.T) {
	// Test various database error scenarios
	testCases := []struct {
		name            string
		dbError         error
		shouldBeError   bool
		expectedBalance float32
	}{
		{
			name:            "no rows found",
			dbError:         sql.ErrNoRows,
			shouldBeError:   false,
			expectedBalance: 0,
		},
		{
			name:            "connection timeout",
			dbError:         errors.New("connection timeout"),
			shouldBeError:   true,
			expectedBalance: 0,
		},
		{
			name:            "invalid syntax",
			dbError:         errors.New("syntax error"),
			shouldBeError:   true,
			expectedBalance: 0,
		},
		{
			name:            "table does not exist",
			dbError:         errors.New("table balances does not exist"),
			shouldBeError:   true,
			expectedBalance: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockDB := test.NewMockDatabase()
			repo := NewBalanceRepository(mockDB)

			userId := "test-user"
			expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`

			// Configure mock
			mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
				Return(tc.dbError)

			// Act
			result, err := repo.GetBalance(userId)

			// Assert
			if tc.shouldBeError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get balance for user")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedBalance, result.AvailableBalance)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestBalanceRepository_Integration_Scenarios(t *testing.T) {
	t.Run("multiple consecutive calls", func(t *testing.T) {
		// Arrange
		mockDB := test.NewMockDatabase()
		repo := NewBalanceRepository(mockDB)

		userId := "test-user"
		expectedQuery := `SELECT available_balance FROM balances WHERE user_id = $1`
		expectedBalance := float32(1000.0)

		// Configure mock for multiple calls
		mockDB.On("Get", mock.AnythingOfType("*float32"), expectedQuery, []interface{}{userId}).
			Return(nil, expectedBalance).Times(3)

		// Act & Assert - Multiple calls should work
		for i := 0; i < 3; i++ {
			result, err := repo.GetBalance(userId)
			assert.NoError(t, err)
			assert.Equal(t, expectedBalance, result.AvailableBalance)
		}

		mockDB.AssertExpectations(t)
	})
}
