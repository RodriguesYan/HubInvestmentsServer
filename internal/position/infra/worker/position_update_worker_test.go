package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"HubInvestments/internal/position/application/command"
	domain "HubInvestments/internal/position/domain/model"
	sharedMessaging "HubInvestments/shared/infra/messaging"

	"github.com/google/uuid"
)

// Mock implementations for testing

type MockCreatePositionUseCase struct {
	ExecuteFunc func(ctx context.Context, cmd *command.CreatePositionCommand) (*command.CreatePositionResult, error)
}

func (m *MockCreatePositionUseCase) Execute(ctx context.Context, cmd *command.CreatePositionCommand) (*command.CreatePositionResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, cmd)
	}
	return &command.CreatePositionResult{
		PositionID:      uuid.New().String(),
		Status:          "ACTIVE",
		TotalInvestment: cmd.Quantity * cmd.Price,
		Message:         "Position created successfully",
	}, nil
}

type MockUpdatePositionUseCase struct {
	ExecuteFunc func(ctx context.Context, cmd *command.UpdatePositionCommand) (*command.UpdatePositionResult, error)
}

func (m *MockUpdatePositionUseCase) Execute(ctx context.Context, cmd *command.UpdatePositionCommand) (*command.UpdatePositionResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, cmd)
	}
	return &command.UpdatePositionResult{
		PositionID:         cmd.PositionID,
		NewQuantity:        100.0,
		NewAveragePrice:    150.0,
		NewTotalInvestment: 15000.0,
		Status:             "ACTIVE",
		TransactionType:    "BUY",
		Message:            "Position updated successfully",
	}, nil
}

type MockClosePositionUseCase struct {
	ExecuteFunc func(ctx context.Context, cmd *command.ClosePositionCommand) (*command.ClosePositionResult, error)
}

func (m *MockClosePositionUseCase) Execute(ctx context.Context, cmd *command.ClosePositionCommand) (*command.ClosePositionResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, cmd)
	}
	return &command.ClosePositionResult{
		UpdatePositionResult: &command.UpdatePositionResult{
			PositionID:         cmd.PositionID,
			NewQuantity:        0.0,
			NewAveragePrice:    150.0,
			NewTotalInvestment: 0.0,
			Status:             "CLOSED",
			TransactionType:    "SELL",
			Message:            "Position closed successfully",
		},
		HoldingPeriodDays:  30.5,
		TotalRealizedValue: 15500.0,
		FinalSellPrice:     155.0,
		PositionClosedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

type MockPositionRepository struct {
	ExistsForUserFunc         func(ctx context.Context, userID uuid.UUID, symbol string) (bool, error)
	FindByUserIDFunc          func(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error)
	FindByUserIDAndSymbolFunc func(ctx context.Context, userID uuid.UUID, symbol string) (*domain.Position, error)
	FindActivePositionsFunc   func(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error)
}

func (m *MockPositionRepository) ExistsForUser(ctx context.Context, userID uuid.UUID, symbol string) (bool, error) {
	if m.ExistsForUserFunc != nil {
		return m.ExistsForUserFunc(ctx, userID, symbol)
	}
	return false, nil
}

func (m *MockPositionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	if m.FindByUserIDFunc != nil {
		return m.FindByUserIDFunc(ctx, userID)
	}
	return []*domain.Position{}, nil
}

func (m *MockPositionRepository) FindByUserIDAndSymbol(ctx context.Context, userID uuid.UUID, symbol string) (*domain.Position, error) {
	if m.FindByUserIDAndSymbolFunc != nil {
		return m.FindByUserIDAndSymbolFunc(ctx, userID, symbol)
	}
	return nil, nil
}

func (m *MockPositionRepository) FindActivePositions(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	if m.FindActivePositionsFunc != nil {
		return m.FindActivePositionsFunc(ctx, userID)
	}
	return []*domain.Position{}, nil
}

func (m *MockPositionRepository) FindByID(ctx context.Context, positionID uuid.UUID) (*domain.Position, error) {
	return nil, nil
}

func (m *MockPositionRepository) Save(ctx context.Context, position *domain.Position) error {
	return nil
}

func (m *MockPositionRepository) Update(ctx context.Context, position *domain.Position) error {
	return nil
}

func (m *MockPositionRepository) Delete(ctx context.Context, positionID uuid.UUID) error {
	return nil
}

func (m *MockPositionRepository) CountPositionsForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *MockPositionRepository) GetTotalInvestmentForUser(ctx context.Context, userID uuid.UUID) (float64, error) {
	return 0.0, nil
}

type MockMessageHandler struct {
	PublishWithOptionsFunc func(ctx context.Context, options sharedMessaging.PublishOptions) error
}

func (m *MockMessageHandler) Publish(ctx context.Context, queueName string, message []byte) error {
	return nil
}

func (m *MockMessageHandler) PublishWithOptions(ctx context.Context, options sharedMessaging.PublishOptions) error {
	if m.PublishWithOptionsFunc != nil {
		return m.PublishWithOptionsFunc(ctx, options)
	}
	return nil
}

func (m *MockMessageHandler) Consume(ctx context.Context, queueName string, handler sharedMessaging.MessageConsumer) error {
	return nil
}

func (m *MockMessageHandler) DeclareQueue(name string, options sharedMessaging.QueueOptions) error {
	return nil
}

func (m *MockMessageHandler) DeleteQueue(queueName string) error {
	return nil
}

func (m *MockMessageHandler) PurgeQueue(queueName string) error {
	return nil
}

func (m *MockMessageHandler) QueueInfo(queueName string) (*sharedMessaging.QueueInfo, error) {
	return &sharedMessaging.QueueInfo{
		Name:     queueName,
		Messages: 0,
	}, nil
}

func (m *MockMessageHandler) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockMessageHandler) Close() error {
	return nil
}

func TestNewPositionUpdateWorker(t *testing.T) {
	createUC := &MockCreatePositionUseCase{}
	updateUC := &MockUpdatePositionUseCase{}
	closeUC := &MockClosePositionUseCase{}
	positionRepo := &MockPositionRepository{}
	messageHandler := &MockMessageHandler{}

	worker := NewPositionUpdateWorker(
		"test-worker-1",
		createUC,
		updateUC,
		closeUC,
		positionRepo,
		messageHandler,
		nil, // Use default config
	)

	if worker == nil {
		t.Fatal("Expected worker to be created, got nil")
	}

	if worker.GetID() != "test-worker-1" {
		t.Errorf("Expected worker ID 'test-worker-1', got '%s'", worker.GetID())
	}

	if worker.IsRunning() {
		t.Error("Expected worker to not be running initially")
	}

	if worker.GetHealthStatus() != HealthStatusUnknown {
		t.Errorf("Expected health status to be Unknown initially, got %s", worker.GetHealthStatus())
	}
}

func TestPositionWorkerConfig(t *testing.T) {
	config := DefaultPositionWorkerConfig("test-worker")

	if config.WorkerID != "test-worker" {
		t.Errorf("Expected WorkerID 'test-worker', got '%s'", config.WorkerID)
	}

	if config.MaxConcurrentUpdates != 20 {
		t.Errorf("Expected MaxConcurrentUpdates 20, got %d", config.MaxConcurrentUpdates)
	}

	if config.ProcessingTimeout != 15*time.Second {
		t.Errorf("Expected ProcessingTimeout 15s, got %v", config.ProcessingTimeout)
	}

	if config.MaxRetries != 4 {
		t.Errorf("Expected MaxRetries 4, got %d", config.MaxRetries)
	}
}

func TestPositionUpdateWorker_HandleBuyOrder_CreateNewPosition(t *testing.T) {
	createUC := &MockCreatePositionUseCase{}
	updateUC := &MockUpdatePositionUseCase{}
	closeUC := &MockClosePositionUseCase{}
	positionRepo := &MockPositionRepository{
		ExistsForUserFunc: func(ctx context.Context, userID uuid.UUID, symbol string) (bool, error) {
			return false, nil // No existing position
		},
	}
	messageHandler := &MockMessageHandler{}

	worker := NewPositionUpdateWorker(
		"test-worker",
		createUC,
		updateUC,
		closeUC,
		positionRepo,
		messageHandler,
		nil,
	)

	message := &PositionUpdateMessage{
		OrderID:        uuid.New().String(),
		UserID:         uuid.New().String(),
		Symbol:         "AAPL",
		OrderSide:      "BUY",
		Quantity:       100.0,
		ExecutionPrice: 150.0,
		TotalValue:     15000.0,
		ExecutedAt:     time.Now(),
	}

	operationType, err := worker.handleBuyOrder(context.Background(), message)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if operationType != "position_create" {
		t.Errorf("Expected operation type 'position_create', got '%s'", operationType)
	}
}

func TestPositionUpdateWorker_IsRetryableError(t *testing.T) {
	worker := &PositionUpdateWorker{}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection error",
			err:      fmt.Errorf("connection failed"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      fmt.Errorf("request timeout"),
			expected: true,
		},
		{
			name:     "network error",
			err:      fmt.Errorf("network unavailable"),
			expected: true,
		},
		{
			name:     "business logic error",
			err:      fmt.Errorf("invalid position"),
			expected: false,
		},
		{
			name:     "validation error",
			err:      fmt.Errorf("user ID is required"),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := worker.isRetryableError(test.err)
			if result != test.expected {
				t.Errorf("Expected %t for error '%v', got %t", test.expected, test.err, result)
			}
		})
	}
}

func TestPositionWorkerMetrics(t *testing.T) {
	metrics := NewPositionWorkerMetrics()

	if metrics.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}

	if metrics.LastActivityTime.IsZero() {
		t.Error("Expected LastActivityTime to be set")
	}

	if metrics.PositionsProcessed != 0 {
		t.Errorf("Expected PositionsProcessed to be 0, got %d", metrics.PositionsProcessed)
	}
}

func TestHealthStatus_String(t *testing.T) {
	tests := []struct {
		status   HealthStatus
		expected string
	}{
		{HealthStatusHealthy, "healthy"},
		{HealthStatusDegraded, "degraded"},
		{HealthStatusUnhealthy, "unhealthy"},
		{HealthStatusStopped, "stopped"},
		{HealthStatusUnknown, "unknown"},
	}

	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("Expected %s for status %d, got %s", test.expected, int(test.status), result)
		}
	}
}
