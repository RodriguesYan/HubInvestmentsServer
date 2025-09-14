package http

import (
	usecase "HubInvestments/internal/position/application/usecase"
	domain "HubInvestments/internal/position/domain/model"
	di "HubInvestments/pck"
	"HubInvestments/shared/middleware"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"math"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocking the auth package
type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) VerifyToken(token string, w http.ResponseWriter) (string, error) {
	args := m.Called(token, w)
	return args.String(0), args.Error(1)
}

type MockPositionRepository struct {
	positions  []*domain.Position
	categories map[string]int // Store category mapping by symbol
	err        error
}

// Convert AssetModel data to Position domain models for compatibility
func (m *MockPositionRepository) addAssetModels(assets []domain.AssetModel, userUUID uuid.UUID) {
	m.categories = make(map[string]int)
	validPositions := []*domain.Position{}

	for _, asset := range assets {
		// Skip assets with zero quantity as they can't create valid positions
		if asset.Quantity <= 0 {
			// Still store category mapping for consistency
			m.categories[asset.Symbol] = asset.Category
			continue
		}

		position, err := domain.NewPosition(userUUID, asset.Symbol, float64(asset.Quantity), float64(asset.AveragePrice), domain.PositionTypeLong)
		if err != nil {
			// Skip invalid positions
			continue
		}
		position.CurrentPrice = float64(asset.LastPrice)
		position.MarketValue = position.Quantity * position.CurrentPrice
		validPositions = append(validPositions, position)
		m.categories[asset.Symbol] = asset.Category // Store category mapping
	}

	m.positions = validPositions
}

func (m *MockPositionRepository) FindByID(ctx context.Context, positionID uuid.UUID) (*domain.Position, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, position := range m.positions {
		if position.ID == positionID {
			return position, nil
		}
	}
	return nil, nil
}

func (m *MockPositionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	if m.err != nil {
		return nil, m.err
	}
	var userPositions []*domain.Position
	for _, position := range m.positions {
		if position.UserID == userID {
			userPositions = append(userPositions, position)
		}
	}
	return userPositions, nil
}

func (m *MockPositionRepository) FindByUserIDAndSymbol(ctx context.Context, userID uuid.UUID, symbol string) (*domain.Position, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, position := range m.positions {
		if position.UserID == userID && position.Symbol == symbol {
			return position, nil
		}
	}
	return nil, nil
}

func (m *MockPositionRepository) FindActivePositions(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	if m.err != nil {
		return nil, m.err
	}
	var activePositions []*domain.Position
	for _, position := range m.positions {
		if position.UserID == userID && position.Status.CanBeUpdated() {
			activePositions = append(activePositions, position)
		}
	}
	return activePositions, nil
}

func (m *MockPositionRepository) Save(ctx context.Context, position *domain.Position) error {
	if m.err != nil {
		return m.err
	}
	m.positions = append(m.positions, position)
	return nil
}

func (m *MockPositionRepository) Update(ctx context.Context, position *domain.Position) error {
	if m.err != nil {
		return m.err
	}
	for i, pos := range m.positions {
		if pos.ID == position.ID {
			m.positions[i] = position
			return nil
		}
	}
	return errors.New("position not found")
}

func (m *MockPositionRepository) Delete(ctx context.Context, positionID uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	for i, position := range m.positions {
		if position.ID == positionID {
			m.positions = append(m.positions[:i], m.positions[i+1:]...)
			return nil
		}
	}
	return errors.New("position not found")
}

func (m *MockPositionRepository) ExistsForUser(ctx context.Context, userID uuid.UUID, symbol string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	for _, position := range m.positions {
		if position.UserID == userID && position.Symbol == symbol {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockPositionRepository) CountPositionsForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	count := 0
	for _, position := range m.positions {
		if position.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockPositionRepository) GetTotalInvestmentForUser(ctx context.Context, userID uuid.UUID) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	total := 0.0
	for _, position := range m.positions {
		if position.UserID == userID && position.Status.CanBeUpdated() {
			total += position.TotalInvestment
		}
	}
	return total, nil
}

// GetCategoryForSymbol returns the category for a given symbol
func (m *MockPositionRepository) GetCategoryForSymbol(symbol string) int {
	if category, exists := m.categories[symbol]; exists {
		return category
	}
	return 1 // Default category
}

// MockPositionAggregationUseCase for testing JSON marshal errors
type MockPositionAggregationUseCase struct {
	result domain.AucAggregationModel
	err    error
}

func (m *MockPositionAggregationUseCase) Execute(userId string) (domain.AucAggregationModel, error) {
	return m.result, m.err
}

// MockContainer for testing
type MockContainer struct {
	positionUseCase *MockPositionAggregationUseCase
}

func (m *MockContainer) GetAucService() interface{} {
	return nil
}

func (m *MockContainer) GetPositionAggregationUseCase() *MockPositionAggregationUseCase {
	return m.positionUseCase
}

func (m *MockContainer) GetBalanceUseCase() interface{} {
	return nil
}

func (m *MockContainer) GetPortfolioSummaryUsecase() interface{} {
	return nil
}

// MockPositionAggregationUseCaseForJSONError returns data that cannot be JSON marshaled
type MockPositionAggregationUseCaseForJSONError struct{}

func (m *MockPositionAggregationUseCaseForJSONError) Execute(userId string) (domain.AucAggregationModel, error) {
	// Create a response that contains unmarshalable data
	// Since AucAggregationModel doesn't contain channels, we need a different approach
	// Let's create an invalid float value instead
	return domain.AucAggregationModel{
		TotalInvested:       float32(math.Inf(1)), // This should cause JSON marshal to fail
		CurrentTotal:        float32(math.NaN()),  // NaN values can't be marshaled to JSON
		PositionAggregation: []domain.PositionAggregationModel{},
	}, nil
}

func TestGetAucAggregation_Success(t *testing.T) {
	// Mock dependencies
	testUUID := uuid.New()
	expectedUserId := testUUID.String()

	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{
		{Symbol: "AAPL", Category: 1, AveragePrice: 150, LastPrice: 155, Quantity: 10},
		{Symbol: "AMZN", Category: 1, AveragePrice: 350, LastPrice: 385, Quantity: 5},
		{Symbol: "VOO", Category: 2, AveragePrice: 450, LastPrice: 555, Quantity: 15},
	}
	mockRepo.addAssetModels(assets, testUUID)

	// Use the test-specific use case that preserves categories
	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	// Test the direct handler (without middleware authentication)
	GetAucAggregation(rr, req, expectedUserId, testContainer)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response domain.AucAggregationModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// All assets are now in category 1 (single aggregation)
	assert.Equal(t, 1, len(response.PositionAggregation))
	assert.Equal(t, float32(11800), response.PositionAggregation[0].CurrentTotal)
	assert.Equal(t, float32(10000), response.PositionAggregation[0].TotalInvested)
	assert.Equal(t, float32(1800), response.PositionAggregation[0].Pnl)
	assert.Equal(t, float32(18), response.PositionAggregation[0].PnlPercentage)
	assert.Equal(t, int(3), len(response.PositionAggregation[0].Assets))

	// Verify total values
	assert.Equal(t, float32(10000), response.TotalInvested)
	assert.Equal(t, float32(11800), response.CurrentTotal)
}

func TestGetAucAggregation_UseCaseError(t *testing.T) {
	// Mock dependencies with error
	testUUID := uuid.New()
	expectedUserId := testUUID.String()

	mockRepo := &MockPositionRepository{}
	mockRepo.err = errors.New("database connection failed")

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	GetAucAggregation(rr, req, expectedUserId, testContainer)

	// Check error response
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to get position aggregation")
	assert.Contains(t, rr.Body.String(), "database connection failed")
}

func TestGetAucAggregation_EmptyPositions(t *testing.T) {
	// Test with empty positions
	testUUID := uuid.New()
	expectedUserId := testUUID.String()

	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{} // Empty slice
	mockRepo.addAssetModels(assets, testUUID)

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	GetAucAggregation(rr, req, expectedUserId, testContainer)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response domain.AucAggregationModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Should have empty aggregations
	assert.Equal(t, float32(0), response.TotalInvested)
	assert.Equal(t, float32(0), response.CurrentTotal)
	assert.Equal(t, 0, len(response.PositionAggregation))
}

func TestGetAucAggregation_SinglePosition(t *testing.T) {
	// Test with single position
	testUUID := uuid.New()
	expectedUserId := testUUID.String()

	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{
		{Symbol: "AAPL", Category: 1, AveragePrice: 150, LastPrice: 155, Quantity: 10},
	}
	mockRepo.addAssetModels(assets, testUUID)

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	GetAucAggregation(rr, req, expectedUserId, testContainer)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response domain.AucAggregationModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Should have one aggregation with one asset
	assert.Equal(t, 1, len(response.PositionAggregation))
	assert.Equal(t, 1, len(response.PositionAggregation[0].Assets))
	assert.Equal(t, "AAPL", response.PositionAggregation[0].Assets[0].Symbol)
	assert.Equal(t, float32(1500), response.TotalInvested)
	assert.Equal(t, float32(1550), response.CurrentTotal)
}

func TestGetAucAggregationWithAuth_Success(t *testing.T) {
	// Mock dependencies
	testUUID := uuid.New()
	expectedUserId := testUUID.String()
	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return expectedUserId, nil
	})

	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{
		{Symbol: "AAPL", Category: 1, AveragePrice: 150, LastPrice: 155, Quantity: 10},
		{Symbol: "AMZN", Category: 1, AveragePrice: 350, LastPrice: 385, Quantity: 5},
		{Symbol: "VOO", Category: 2, AveragePrice: 450, LastPrice: 555, Quantity: 15},
	}
	mockRepo.addAssetModels(assets, testUUID)

	// Use the reusable TestContainer with the new use case
	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer token")

	rr := httptest.NewRecorder()
	// Test the middleware-wrapped handler
	handler := GetAucAggregationWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response domain.AucAggregationModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// All assets are now in category 1 (single aggregation)
	assert.Equal(t, 1, len(response.PositionAggregation))
	assert.Equal(t, float32(11800), response.PositionAggregation[0].CurrentTotal)
	assert.Equal(t, float32(10000), response.PositionAggregation[0].TotalInvested)
	assert.Equal(t, float32(1800), response.PositionAggregation[0].Pnl)
	assert.Equal(t, float32(18), response.PositionAggregation[0].PnlPercentage)
	assert.Equal(t, int(3), len(response.PositionAggregation[0].Assets))

	// Verify total values
	assert.Equal(t, float32(10000), response.TotalInvested)
	assert.Equal(t, float32(11800), response.CurrentTotal)
}

func TestGetAucAggregationWithAuth_AuthenticationFailure(t *testing.T) {
	// Mock authentication failure
	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return "", errors.New("invalid token")
	})

	testUUID := uuid.New()
	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{}
	mockRepo.addAssetModels(assets, testUUID)

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rr := httptest.NewRecorder()
	handler := GetAucAggregationWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Check error response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid token")
}

func TestGetAucAggregationWithAuth_MissingAuthHeader(t *testing.T) {
	// Mock authentication failure for missing auth header
	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		if token == "" {
			return "", errors.New("missing authorization header")
		}
		return "user123", nil
	})

	testUUID := uuid.New()
	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{}
	mockRepo.addAssetModels(assets, testUUID)

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)
	// No Authorization header set

	rr := httptest.NewRecorder()
	handler := GetAucAggregationWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Check error response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing authorization header")
}

func TestGetAucAggregationWithAuth_UseCaseAndAuthenticationErrors(t *testing.T) {
	// Test case where auth succeeds but use case fails
	testUUID := uuid.New()
	expectedUserId := testUUID.String()
	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return expectedUserId, nil
	})

	mockRepo := &MockPositionRepository{}
	mockRepo.err = errors.New("repository error")

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()
	handler := GetAucAggregationWithAuth(verifyToken, testContainer)
	handler(rr, req)

	// Check error response
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to get position aggregation")
	assert.Contains(t, rr.Body.String(), "repository error")
}

func TestGetAucAggregation_EdgeCaseWithZeroValues(t *testing.T) {
	// Test with zero price/quantity values
	testUUID := uuid.New()
	expectedUserId := testUUID.String()

	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{
		{Symbol: "FREE", Category: 1, AveragePrice: 0, LastPrice: 0, Quantity: 0},
		{Symbol: "ZERO", Category: 1, AveragePrice: 100, LastPrice: 50, Quantity: 0},
	}
	mockRepo.addAssetModels(assets, testUUID)

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	GetAucAggregation(rr, req, expectedUserId, testContainer)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response domain.AucAggregationModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Should handle zero values correctly - no valid positions means no aggregations
	assert.Equal(t, float32(0), response.TotalInvested)
	assert.Equal(t, float32(0), response.CurrentTotal)
	assert.Equal(t, 0, len(response.PositionAggregation))
}

func TestGetAucAggregation_MultipleCategories(t *testing.T) {
	// Test with multiple categories (stocks, bonds, ETFs, etc.)
	testUUID := uuid.New()
	expectedUserId := testUUID.String()

	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{
		{Symbol: "AAPL", Category: 1, AveragePrice: 150, LastPrice: 155, Quantity: 10}, // Stock
		{Symbol: "BOND", Category: 3, AveragePrice: 100, LastPrice: 105, Quantity: 20}, // Bond
		{Symbol: "VOO", Category: 2, AveragePrice: 450, LastPrice: 455, Quantity: 5},   // ETF
		{Symbol: "MSFT", Category: 1, AveragePrice: 300, LastPrice: 310, Quantity: 3},  // Stock
	}
	mockRepo.addAssetModels(assets, testUUID)

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	GetAucAggregation(rr, req, expectedUserId, testContainer)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response domain.AucAggregationModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// All assets are now in category 1 (single aggregation)
	assert.Equal(t, 1, len(response.PositionAggregation))

	// Calculate expected totals - all assets combined
	expectedTotalInvested := float32(150*10 + 300*3 + 450*5 + 100*20) // 1500 + 900 + 2250 + 2000 = 6650
	expectedCurrentTotal := float32(155*10 + 310*3 + 455*5 + 105*20)  // 1550 + 930 + 2275 + 2100 = 6855

	assert.Equal(t, expectedTotalInvested, response.TotalInvested)
	assert.Equal(t, expectedCurrentTotal, response.CurrentTotal)

	// All 4 assets should be in the single aggregation
	assert.Equal(t, 4, len(response.PositionAggregation[0].Assets))
}

// Test HTTP methods and request validation
func TestGetAucAggregation_HTTPMethods(t *testing.T) {
	testUUID := uuid.New()
	expectedUserId := testUUID.String()
	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{}
	mockRepo.addAssetModels(assets, testUUID)

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	// Test GET method (should work)
	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	GetAucAggregation(rr, req, expectedUserId, testContainer)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test POST method (should still work as the handler doesn't check method)
	req, err = http.NewRequest("POST", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	GetAucAggregation(rr, req, expectedUserId, testContainer)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetAucAggregation_UserIdVariations(t *testing.T) {
	// Test different userId formats
	testCases := []struct {
		name           string
		userId         string
		expectedStatus int
	}{
		{"normal user id", "user123", http.StatusInternalServerError},             // Invalid UUID
		{"uuid style", "550e8400-e29b-41d4-a716-446655440000", http.StatusOK},     // Valid UUID
		{"empty user id", "", http.StatusInternalServerError},                     // Invalid UUID
		{"numeric user id", "12345", http.StatusInternalServerError},              // Invalid UUID
		{"special characters", "user@domain.com", http.StatusInternalServerError}, // Invalid UUID
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testUUID := uuid.New()
			mockRepo := &MockPositionRepository{}
			assets := []domain.AssetModel{}
			mockRepo.addAssetModels(assets, testUUID)

			positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
			testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

			req, err := http.NewRequest("GET", "/auc-aggregation", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			GetAucAggregation(rr, req, tc.userId, testContainer)

			// Expect different status codes based on UUID validity
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestGetAucAggregation_JSONMarshalError(t *testing.T) {
	// Test JSON marshaling error by creating invalid float values that can't be marshaled
	testUUID := uuid.New()
	expectedUserId := testUUID.String()

	// Create mock repository that returns data with invalid float values
	mockRepo := &MockPositionRepository{}
	assets := []domain.AssetModel{
		{
			Symbol:       "TEST",
			Category:     1,
			AveragePrice: float32(math.Inf(1)), // Infinity can't be marshaled to JSON
			LastPrice:    float32(math.NaN()),  // NaN can't be marshaled to JSON
			Quantity:     1,
		},
	}
	mockRepo.addAssetModels(assets, testUUID)

	positionUseCase := usecase.NewGetPositionAggregationUseCase(mockRepo)
	testContainer := di.NewTestContainer().WithPositionAggregationUseCase(positionUseCase)

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	GetAucAggregation(rr, req, expectedUserId, testContainer)

	// Check that it returns an error due to JSON marshaling failure
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	// The error message should contain JSON error
	errorMessage := rr.Body.String()
	assert.NotEmpty(t, errorMessage)
}
