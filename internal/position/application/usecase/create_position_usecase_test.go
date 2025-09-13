package usecase

import (
	"context"
	"testing"

	"HubInvestments/internal/position/application/command"

	"github.com/google/uuid"
)

func TestCreatePositionUseCase_Execute_Success(t *testing.T) {
	// Setup
	mockRepo := NewMockPositionRepositoryForNew()
	usecase := NewCreatePositionUseCase(mockRepo)

	userID := uuid.New()
	cmd := &command.CreatePositionCommand{
		UserID:       userID.String(),
		Symbol:       "AAPL",
		Quantity:     100.0,
		Price:        150.0,
		PositionType: "LONG",
		CreatedFrom:  "MANUAL_ENTRY",
	}

	// Execute
	result, err := usecase.Execute(context.Background(), cmd)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Status != "ACTIVE" {
		t.Errorf("Expected status ACTIVE, got: %s", result.Status)
	}

	expectedInvestment := 100.0 * 150.0
	if result.TotalInvestment != expectedInvestment {
		t.Errorf("Expected total investment %.2f, got: %.2f", expectedInvestment, result.TotalInvestment)
	}

	// Verify position was saved
	savedPositions := len(mockRepo.positions)
	if savedPositions != 1 {
		t.Errorf("Expected 1 saved position, got: %d", savedPositions)
	}
}

func TestCreatePositionUseCase_Execute_ValidationError(t *testing.T) {
	// Setup
	mockRepo := NewMockPositionRepositoryForNew()
	usecase := NewCreatePositionUseCase(mockRepo)

	tests := []struct {
		name        string
		cmd         *command.CreatePositionCommand
		expectedErr string
	}{
		{
			name: "Empty UserID",
			cmd: &command.CreatePositionCommand{
				UserID:       "",
				Symbol:       "AAPL",
				Quantity:     100.0,
				Price:        150.0,
				PositionType: "LONG",
			},
			expectedErr: "invalid command",
		},
		{
			name: "Invalid UserID format",
			cmd: &command.CreatePositionCommand{
				UserID:       "invalid-uuid",
				Symbol:       "AAPL",
				Quantity:     100.0,
				Price:        150.0,
				PositionType: "LONG",
			},
			expectedErr: "invalid command",
		},
		{
			name: "Empty Symbol",
			cmd: &command.CreatePositionCommand{
				UserID:       uuid.New().String(),
				Symbol:       "",
				Quantity:     100.0,
				Price:        150.0,
				PositionType: "LONG",
			},
			expectedErr: "invalid command",
		},
		{
			name: "Zero Quantity",
			cmd: &command.CreatePositionCommand{
				UserID:       uuid.New().String(),
				Symbol:       "AAPL",
				Quantity:     0.0,
				Price:        150.0,
				PositionType: "LONG",
			},
			expectedErr: "invalid command",
		},
		{
			name: "Zero Price",
			cmd: &command.CreatePositionCommand{
				UserID:       uuid.New().String(),
				Symbol:       "AAPL",
				Quantity:     100.0,
				Price:        0.0,
				PositionType: "LONG",
			},
			expectedErr: "invalid command",
		},
		{
			name: "Invalid Position Type",
			cmd: &command.CreatePositionCommand{
				UserID:       uuid.New().String(),
				Symbol:       "AAPL",
				Quantity:     100.0,
				Price:        150.0,
				PositionType: "INVALID",
			},
			expectedErr: "invalid command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := usecase.Execute(context.Background(), tt.cmd)

			// Assert
			if err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.expectedErr)
			}

			if result != nil {
				t.Errorf("Expected nil result, got: %v", result)
			}
		})
	}
}

func TestCreatePositionUseCase_Execute_PositionAlreadyExists(t *testing.T) {
	// Setup
	mockRepo := NewMockPositionRepositoryForNew()
	usecase := NewCreatePositionUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.SetExistsForUser(userID, "AAPL", true)

	cmd := &command.CreatePositionCommand{
		UserID:       userID.String(),
		Symbol:       "AAPL",
		Quantity:     100.0,
		Price:        150.0,
		PositionType: "LONG",
	}

	// Execute
	result, err := usecase.Execute(context.Background(), cmd)

	// Assert
	if err == nil {
		t.Error("Expected error for existing position, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	expectedErrMsg := "position already exists"
	if err != nil && len(err.Error()) > 0 {
		if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
			t.Errorf("Expected error starting with '%s', got: %s", expectedErrMsg, err.Error())
		}
	}
}

func TestCreatePositionUseCase_Execute_RepositoryError(t *testing.T) {
	// Setup
	mockRepo := NewMockPositionRepositoryForNew()
	usecase := NewCreatePositionUseCase(mockRepo)

	userID := uuid.New()
	cmd := &command.CreatePositionCommand{
		UserID:       userID.String(),
		Symbol:       "AAPL",
		Quantity:     100.0,
		Price:        150.0,
		PositionType: "LONG",
	}

	tests := []struct {
		name           string
		setupMock      func(*MockPositionRepositoryForNew)
		expectedErrMsg string
	}{
		{
			name: "Exists check fails",
			setupMock: func(m *MockPositionRepositoryForNew) {
				m.shouldFailExists = true
			},
			expectedErrMsg: "failed to check existing position",
		},
		{
			name: "Save fails",
			setupMock: func(m *MockPositionRepositoryForNew) {
				m.shouldFailSave = true
			},
			expectedErrMsg: "failed to save position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup specific mock behavior
			tt.setupMock(mockRepo)

			// Execute
			result, err := usecase.Execute(context.Background(), cmd)

			// Assert
			if err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.expectedErrMsg)
			}

			if result != nil {
				t.Errorf("Expected nil result, got: %v", result)
			}

			// Reset mock
			mockRepo.shouldFailExists = false
			mockRepo.shouldFailSave = false
		})
	}
}

func TestCreatePositionUseCase_Execute_WithSourceOrderID(t *testing.T) {
	// Setup
	mockRepo := NewMockPositionRepositoryForNew()
	usecase := NewCreatePositionUseCase(mockRepo)

	userID := uuid.New()
	sourceOrderID := uuid.New().String()

	cmd := &command.CreatePositionCommand{
		UserID:        userID.String(),
		Symbol:        "AAPL",
		Quantity:      100.0,
		Price:         150.0,
		PositionType:  "LONG",
		SourceOrderID: &sourceOrderID,
		CreatedFrom:   "ORDER_EXECUTION",
	}

	// Execute
	result, err := usecase.Execute(context.Background(), cmd)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Verify position was saved with correct context
	if len(mockRepo.positions) != 1 {
		t.Errorf("Expected 1 saved position, got: %d", len(mockRepo.positions))
	}
}
