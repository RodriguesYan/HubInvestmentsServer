package usecase

import (
	"errors"
	"testing"

	domain "HubInvestments/internal/balance/domain/model"
)

// MockBalanceRepository is a mock implementation of IBalanceRepository for testing
type MockBalanceRepository struct {
	GetBalanceFunc func(userId string) (domain.BalanceModel, error)
}

func (m *MockBalanceRepository) GetBalance(userId string) (domain.BalanceModel, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(userId)
	}
	return domain.BalanceModel{}, nil
}

func TestNewGetBalanceUseCase(t *testing.T) {
	mockRepo := &MockBalanceRepository{}
	useCase := NewGetBalanceUseCase(mockRepo)

	if useCase == nil {
		t.Fatal("NewGetBalanceUseCase should not return nil")
	}

	if useCase.repo != mockRepo {
		t.Error("NewGetBalanceUseCase should set the repository correctly")
	}
}

func TestGetBalanceUseCase_Execute_Success(t *testing.T) {
	// Arrange
	userId := "user123"
	expectedBalance := domain.BalanceModel{
		AvailableBalance: 15000.50,
	}

	mockRepo := &MockBalanceRepository{
		GetBalanceFunc: func(id string) (domain.BalanceModel, error) {
			if id != userId {
				t.Errorf("Expected userId %s, got %s", userId, id)
			}
			return expectedBalance, nil
		},
	}

	useCase := NewGetBalanceUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(userId)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.AvailableBalance != expectedBalance.AvailableBalance {
		t.Errorf("Expected balance %f, got %f", expectedBalance.AvailableBalance, result.AvailableBalance)
	}
}

func TestGetBalanceUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	userId := "user123"
	expectedError := errors.New("database connection failed")

	mockRepo := &MockBalanceRepository{
		GetBalanceFunc: func(id string) (domain.BalanceModel, error) {
			return domain.BalanceModel{}, expectedError
		},
	}

	useCase := NewGetBalanceUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(userId)

	// Assert
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}

	// Should return empty balance model on error
	if result.AvailableBalance != 0 {
		t.Errorf("Expected empty balance model on error, got %+v", result)
	}
}
