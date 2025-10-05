package usecase

import (
	domain "HubInvestments/internal/position/domain/model"
	service "HubInvestments/internal/position/domain/service"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_GetPositionAggregationUseCase_Success(t *testing.T) {
	userUUID := uuid.New()
	userId := userUUID.String()

	// Create positions using new domain model
	position1, _ := domain.NewPosition(userUUID, "AAPL", 5.0, 10.0, domain.PositionTypeLong)
	position1.CurrentPrice = 11.0
	position1.MarketValue = position1.Quantity * position1.CurrentPrice

	position2, _ := domain.NewPosition(userUUID, "AAPL", 7.0, 11.0, domain.PositionTypeLong)
	position2.CurrentPrice = 11.0
	position2.MarketValue = position2.Quantity * position2.CurrentPrice

	repo := NewMockPositionRepositoryForNew()
	repo.AddPosition(position1)
	repo.AddPosition(position2)

	usecase, err := NewGetPositionAggregationUseCase(repo).Execute(userId)

	assert.NoError(t, err)

	assert.Equal(t, 1, len(usecase.PositionAggregation))
	assert.Equal(t, 1, usecase.PositionAggregation[0].Category)
	assert.Equal(t, float32(5.0), usecase.PositionAggregation[0].Assets[0].Quantity)
	assert.Equal(t, float32(10.0), usecase.PositionAggregation[0].Assets[0].AveragePrice)
	assert.Equal(t, float32(11.0), usecase.PositionAggregation[0].Assets[0].LastPrice)
	assert.Equal(t, float32(127.0), usecase.PositionAggregation[0].TotalInvested)
	assert.Equal(t, float32(132.0), usecase.PositionAggregation[0].CurrentTotal)
	assert.Equal(t, float32(5), usecase.PositionAggregation[0].Pnl)
	assert.Equal(t, float32(3.937008), usecase.PositionAggregation[0].PnlPercentage)
	assert.Equal(t, float32(127.0), usecase.TotalInvested)
	assert.Equal(t, float32(132.0), usecase.CurrentTotal)

	//print the result of usecase in a formatted way
	fmt.Printf("%+v\n", usecase)
}

func Test_GetPositionAggregationUseCase_FailRepo(t *testing.T) {
	userUUID := uuid.New()
	userId := userUUID.String()

	repo := NewMockPositionRepositoryForNew()
	repo.shouldFailFind = true

	_, err := NewGetPositionAggregationUseCase(repo).Execute(userId)

	assert.Error(t, err)
}

func Test_GetPositionAggregationUseCase_WithDependencyInjection(t *testing.T) {
	userUUID := uuid.New()
	userId := userUUID.String()

	// Create positions using new domain model
	position1, _ := domain.NewPosition(userUUID, "AAPL", 5.0, 10.0, domain.PositionTypeLong)
	position1.CurrentPrice = 11.0
	position1.MarketValue = position1.Quantity * position1.CurrentPrice

	position2, _ := domain.NewPosition(userUUID, "GOOGL", 2.0, 20.0, domain.PositionTypeLong)
	position2.CurrentPrice = 22.0
	position2.MarketValue = position2.Quantity * position2.CurrentPrice

	repo := NewMockPositionRepositoryForNew()
	repo.AddPosition(position1)
	repo.AddPosition(position2)

	// Use the real domain service
	aggregationService := service.NewPositionAggregationService()

	// Create use case with dependency injection
	useCaseWithDI := NewGetPositionAggregationUseCaseWithService(repo, aggregationService)
	result, err := useCaseWithDI.Execute(userId)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.PositionAggregation))
	assert.Equal(t, 1, result.PositionAggregation[0].Category)
	assert.Equal(t, float32(90.0), result.PositionAggregation[0].TotalInvested) // (5*10) + (2*20) = 50 + 40 = 90
	assert.Equal(t, float32(99.0), result.PositionAggregation[0].CurrentTotal)  // (5*11) + (2*22) = 55 + 44 = 99
	assert.Equal(t, float32(9.0), result.PositionAggregation[0].Pnl)            // 99 - 90 = 9
	assert.Equal(t, float32(90.0), result.TotalInvested)
	assert.Equal(t, float32(99.0), result.CurrentTotal)
	assert.Len(t, result.PositionAggregation[0].Assets, 2)
}

func Test_GetPositionAggregationUseCase_InvalidUserID(t *testing.T) {
	invalidUserId := "invalid-user-id-format"

	repo := NewMockPositionRepositoryForNew()

	_, err := NewGetPositionAggregationUseCase(repo).Execute(invalidUserId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID format")
	assert.Contains(t, err.Error(), "cannot be parsed as UUID or integer")
}

func Test_GetPositionAggregationUseCase_EmptyPositions(t *testing.T) {
	userUUID := uuid.New()
	userId := userUUID.String()

	repo := NewMockPositionRepositoryForNew()

	result, err := NewGetPositionAggregationUseCase(repo).Execute(userId)

	assert.NoError(t, err)
	assert.Equal(t, float32(0.0), result.TotalInvested)
	assert.Equal(t, float32(0.0), result.CurrentTotal)
	assert.Len(t, result.PositionAggregation, 0)
}

func TestParseUserIDToUUID(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		expectError bool
		description string
	}{
		{
			name:        "Valid UUID string",
			userID:      "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
			description: "Should parse valid UUID string directly",
		},
		{
			name:        "Integer string - single digit",
			userID:      "1",
			expectError: false,
			description: "Should convert integer '1' to deterministic UUID",
		},
		{
			name:        "Integer string - multiple digits",
			userID:      "12345",
			expectError: false,
			description: "Should convert integer '12345' to deterministic UUID",
		},
		{
			name:        "Invalid string",
			userID:      "invalid-user-id",
			expectError: true,
			description: "Should fail for non-UUID, non-integer strings",
		},
		{
			name:        "Empty string",
			userID:      "",
			expectError: true,
			description: "Should fail for empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseUserIDToUUID(tt.userID)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Equal(t, uuid.Nil, result, "Should return Nil UUID on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEqual(t, uuid.Nil, result, "Should return valid UUID")

				// Test consistency: same input should produce same UUID
				result2, err2 := parseUserIDToUUID(tt.userID)
				assert.NoError(t, err2)
				assert.Equal(t, result, result2, "Same input should produce consistent UUID")
			}
		})
	}
}

func Test_GetPositionAggregationUseCase_IntegerUserID(t *testing.T) {
	// This test specifically addresses the original issue: userId="1"
	integerUserId := "1"

	// The parseUserIDToUUID should convert "1" to a deterministic UUID
	expectedUUID, err := parseUserIDToUUID(integerUserId)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, expectedUUID)

	// Create a position with the converted UUID
	position1, _ := domain.NewPosition(expectedUUID, "AAPL", 5.0, 10.0, domain.PositionTypeLong)
	position1.CurrentPrice = 11.0
	position1.MarketValue = position1.Quantity * position1.CurrentPrice

	repo := NewMockPositionRepositoryForNew()
	repo.AddPosition(position1)

	// Execute with string "1" - should work now
	result, err := NewGetPositionAggregationUseCase(repo).Execute(integerUserId)

	assert.NoError(t, err, "Should successfully handle integer user ID '1'")
	assert.Equal(t, 1, len(result.PositionAggregation))
	assert.Equal(t, float32(50.0), result.TotalInvested) // 5 * 10
	assert.Equal(t, float32(55.0), result.CurrentTotal)  // 5 * 11
}
