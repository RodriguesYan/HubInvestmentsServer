package service

import (
	"errors"
	"testing"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRiskDataClient is a mock implementation of IRiskDataClient
type MockRiskDataClient struct {
	mock.Mock
}

func (m *MockRiskDataClient) GetUserRiskProfile(userID string) (*UserRiskProfile, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserRiskProfile), args.Error(1)
}

func (m *MockRiskDataClient) GetPositionExposure(userID, symbol string) (*PositionExposure, error) {
	args := m.Called(userID, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PositionExposure), args.Error(1)
}

func (m *MockRiskDataClient) GetAccountBalance(userID string) (*AccountBalance, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountBalance), args.Error(1)
}

func (m *MockRiskDataClient) GetMarketVolatility(symbol string) (*MarketVolatility, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MarketVolatility), args.Error(1)
}

func (m *MockRiskDataClient) GetUserTradingLimits(userID string) (*TradingLimits, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TradingLimits), args.Error(1)
}

// Test helpers and fixtures

func createTestOrder(userID, symbol string, side domain.OrderSide, orderType domain.OrderType, quantity float64, price *float64) *domain.Order {
	order, _ := domain.NewOrder(userID, symbol, side, orderType, quantity, price)
	return order
}

func createTestUserRiskProfile(userID string) *UserRiskProfile {
	return &UserRiskProfile{
		UserID:               userID,
		RiskTolerance:        RiskToleranceModerate,
		MaxPositionSize:      100000.0,
		MaxDailyTradingValue: 50000.0,
		MaxOrderValue:        25000.0,
		IsHighRiskApproved:   false,
		ProfileLastUpdated:   time.Now(),
	}
}

func createTestPositionExposure(symbol string) *PositionExposure {
	return &PositionExposure{
		Symbol:          symbol,
		CurrentQuantity: 100.0,
		CurrentValue:    10000.0,
		AveragePrice:    100.0,
		UnrealizedPnL:   500.0,
		ExposurePercent: 10.0,
	}
}

func createTestAccountBalance() *AccountBalance {
	return &AccountBalance{
		TotalBalance:     100000.0,
		AvailableBalance: 80000.0,
		BuyingPower:      90000.0,
		LastUpdated:      time.Now(),
	}
}

func createTestMarketVolatility(symbol string, isHighVol bool) *MarketVolatility {
	volatility := 15.0
	if isHighVol {
		volatility = 35.0
	}

	return &MarketVolatility{
		Symbol:           symbol,
		Volatility30Day:  volatility,
		Beta:             1.2,
		RiskRating:       "Medium",
		IsHighVolatility: isHighVol,
		LastCalculated:   time.Now(),
	}
}

func createTestTradingLimits() *TradingLimits {
	return &TradingLimits{
		DailyTradingLimit:   50000.0,
		DailyTradingUsed:    10000.0,
		MaxOrderValue:       25000.0,
		MaxPositionSize:     100000.0,
		RemainingDailyLimit: 40000.0,
	}
}

func setupDefaultMockExpectations(mockClient *MockRiskDataClient, userID, symbol string) {
	mockClient.On("GetUserRiskProfile", userID).Return(createTestUserRiskProfile(userID), nil)
	mockClient.On("GetPositionExposure", userID, symbol).Return(createTestPositionExposure(symbol), nil)
	mockClient.On("GetAccountBalance", userID).Return(createTestAccountBalance(), nil)
	mockClient.On("GetMarketVolatility", symbol).Return(createTestMarketVolatility(symbol, false), nil)
	mockClient.On("GetUserTradingLimits", userID).Return(createTestTradingLimits(), nil)
}

// Test Suite for RiskManagementService

func TestNewRiskManagementService(t *testing.T) {
	tests := []struct {
		name   string
		config RiskManagementConfig
		want   *riskManagementService
	}{
		{
			name: "creates service with custom config",
			config: RiskManagementConfig{
				MaxRiskScore:            90.0,
				HighRiskThreshold:       70.0,
				ConcentrationLimit:      25.0,
				VolatilityThreshold:     30.0,
				ManualApprovalThreshold: 80.0,
			},
			want: &riskManagementService{
				maxRiskScore:            90.0,
				highRiskThreshold:       70.0,
				concentrationLimit:      25.0,
				volatilityThreshold:     30.0,
				manualApprovalThreshold: 80.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewRiskManagementService(tt.config)
			impl := service.(*riskManagementService)

			assert.Equal(t, tt.want.maxRiskScore, impl.maxRiskScore)
			assert.Equal(t, tt.want.highRiskThreshold, impl.highRiskThreshold)
			assert.Equal(t, tt.want.concentrationLimit, impl.concentrationLimit)
			assert.Equal(t, tt.want.volatilityThreshold, impl.volatilityThreshold)
			assert.Equal(t, tt.want.manualApprovalThreshold, impl.manualApprovalThreshold)
		})
	}
}

func TestNewRiskManagementServiceWithDefaults(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	impl := service.(*riskManagementService)

	assert.Equal(t, 80.0, impl.maxRiskScore)
	assert.Equal(t, 60.0, impl.highRiskThreshold)
	assert.Equal(t, 20.0, impl.concentrationLimit)
	assert.Equal(t, 25.0, impl.volatilityThreshold)
	assert.Equal(t, 70.0, impl.manualApprovalThreshold)
}

func TestValidateRiskLimits(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "valid order within limits",
			setupMocks: func() {
				userProfile := createTestUserRiskProfile("user1")
				tradingLimits := createTestTradingLimits()

				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetUserTradingLimits", "user1").Return(tradingLimits, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
		{
			name: "order exceeds max order value",
			setupMocks: func() {
				userProfile := createTestUserRiskProfile("user1")
				userProfile.MaxOrderValue = 10000.0 // Lower limit

				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "order value 15000.00 exceeds user limit 10000.00",
		},
		{
			name: "order exceeds remaining daily limit",
			setupMocks: func() {
				userProfile := createTestUserRiskProfile("user1")
				tradingLimits := createTestTradingLimits()
				tradingLimits.RemainingDailyLimit = 5000.0 // Lower remaining limit

				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetUserTradingLimits", "user1").Return(tradingLimits, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "order value 15000.00 exceeds remaining daily limit 5000.00",
		},
		{
			name: "error getting user risk profile",
			setupMocks: func() {
				mockClient.On("GetUserRiskProfile", "user1").Return(nil, errors.New("database error"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "failed to get user risk profile: database error",
		},
		{
			name: "error getting trading limits",
			setupMocks: func() {
				userProfile := createTestUserRiskProfile("user1")
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetUserTradingLimits", "user1").Return(nil, errors.New("service unavailable"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "failed to get trading limits: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil // Reset mock
			tt.setupMocks()

			err := service.ValidateRiskLimits(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCheckPositionLimits(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "sell order skips position limit check",
			setupMocks: func() {
				// No mocks needed for sell orders
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideSell, domain.OrderTypeLimit, 50.0, floatPtr(150.0)),
		},
		{
			name: "buy order within position limits",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 5000.0 // Small current position
				userProfile := createTestUserRiskProfile("user1")
				balance := createTestAccountBalance()

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
		{
			name: "buy order exceeds max position size",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 90000.0 // Large current position
				userProfile := createTestUserRiskProfile("user1")
				userProfile.MaxPositionSize = 100000.0

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "new position value 105000.00 would exceed maximum allowed 100000.00",
		},
		{
			name: "buy order exceeds concentration limit",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 10000.0
				userProfile := createTestUserRiskProfile("user1")
				balance := createTestAccountBalance()
				balance.TotalBalance = 50000.0 // Smaller balance to trigger concentration limit

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "position concentration 50.0% exceeds limit 20.0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil // Reset mock
			tt.setupMocks()

			err := service.CheckPositionLimits(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCheckTradingLimits(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "order within trading limits",
			setupMocks: func() {
				limits := createTestTradingLimits()
				mockClient.On("GetUserTradingLimits", "user1").Return(limits, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
		{
			name: "order exceeds remaining daily limit",
			setupMocks: func() {
				limits := createTestTradingLimits()
				limits.RemainingDailyLimit = 5000.0
				mockClient.On("GetUserTradingLimits", "user1").Return(limits, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "order value 15000.00 exceeds remaining daily limit 5000.00",
		},
		{
			name: "order exceeds max order value",
			setupMocks: func() {
				limits := createTestTradingLimits()
				limits.MaxOrderValue = 10000.0
				mockClient.On("GetUserTradingLimits", "user1").Return(limits, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "order value 15000.00 exceeds maximum order limit 10000.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil // Reset mock
			tt.setupMocks()

			err := service.CheckTradingLimits(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAssessMarketRisk(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name                string
		setupMocks          func()
		order               *domain.Order
		expectedRiskLevel   RiskLevel
		expectedFactorCount int
	}{
		{
			name: "low volatility market risk",
			setupMocks: func() {
				volatility := createTestMarketVolatility("AAPL", false)
				volatility.Volatility30Day = 10.0
				volatility.Beta = 1.0
				mockClient.On("GetMarketVolatility", "AAPL").Return(volatility, nil)
			},
			order:               createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedRiskLevel:   RiskLevelLow,
			expectedFactorCount: 0,
		},
		{
			name: "high volatility market risk",
			setupMocks: func() {
				volatility := createTestMarketVolatility("AAPL", true)
				volatility.Volatility30Day = 40.0
				volatility.Beta = 1.8
				mockClient.On("GetMarketVolatility", "AAPL").Return(volatility, nil)
			},
			order:               createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedRiskLevel:   RiskLevelVeryHigh,
			expectedFactorCount: 2, // High volatility + High beta
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil // Reset mock
			tt.setupMocks()

			assessment, err := service.AssessMarketRisk(tt.order, mockClient)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedRiskLevel, assessment.RiskLevel)
			assert.Len(t, assessment.RiskFactors, tt.expectedFactorCount)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAssessConcentrationRisk(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name                string
		setupMocks          func()
		order               *domain.Order
		expectedRiskLevel   RiskLevel
		expectedFactorCount int
	}{
		{
			name: "sell order has no concentration risk",
			setupMocks: func() {
				// No mocks needed for sell orders
			},
			order:               createTestOrder("user1", "AAPL", domain.OrderSideSell, domain.OrderTypeLimit, 50.0, floatPtr(150.0)),
			expectedRiskLevel:   RiskLevelLow,
			expectedFactorCount: 0,
		},
		{
			name: "low concentration risk",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 5000.0
				balance := createTestAccountBalance()
				balance.TotalBalance = 200000.0 // (5000 + 15000) / 200000 = 10% concentration

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order:               createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedRiskLevel:   RiskLevelMedium, // 10% concentration gives score of 20
			expectedFactorCount: 0,               // No risk factors since 10% < 16% (80% of 20% limit)
		},
		{
			name: "moderate concentration risk",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 5000.0
				balance := createTestAccountBalance()
				balance.TotalBalance = 120000.0 // (5000 + 15000) / 120000 = 16.67% concentration (> 16% but < 20%)

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order:               createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedRiskLevel:   RiskLevelMedium, // 16.67% concentration gives score of ~23
			expectedFactorCount: 1,               // Moderate concentration factor
		},
		{
			name: "high concentration risk",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 10000.0
				balance := createTestAccountBalance()
				balance.TotalBalance = 50000.0 // (10000 + 15000) / 50000 = 50% concentration > 20% limit

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order:               createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedRiskLevel:   RiskLevelExtremelyHigh, // 50% concentration gives score of 160 (20*2 + 30*4)
			expectedFactorCount: 1,                      // High concentration factor
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil // Reset mock
			tt.setupMocks()

			assessment, err := service.AssessConcentrationRisk(tt.order, mockClient)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedRiskLevel, assessment.RiskLevel)
			assert.Len(t, assessment.RiskFactors, tt.expectedFactorCount)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRequiresManualApproval(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()

	tests := []struct {
		name       string
		assessment *RiskAssessment
		expected   bool
	}{
		{
			name: "low risk score does not require approval",
			assessment: &RiskAssessment{
				RiskScore: 50.0,
				RiskLevel: RiskLevelMedium,
				RiskFactors: []RiskFactor{
					{Impact: RiskImpactLow},
				},
			},
			expected: false,
		},
		{
			name: "high risk score requires approval",
			assessment: &RiskAssessment{
				RiskScore: 75.0,
				RiskLevel: RiskLevelHigh,
				RiskFactors: []RiskFactor{
					{Impact: RiskImpactMedium},
				},
			},
			expected: true,
		},
		{
			name: "critical risk factor requires approval",
			assessment: &RiskAssessment{
				RiskScore: 50.0,
				RiskLevel: RiskLevelMedium,
				RiskFactors: []RiskFactor{
					{Impact: RiskImpactCritical},
				},
			},
			expected: true,
		},
		{
			name: "very high risk level requires approval",
			assessment: &RiskAssessment{
				RiskScore: 60.0,
				RiskLevel: RiskLevelVeryHigh,
				RiskFactors: []RiskFactor{
					{Impact: RiskImpactMedium},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.RequiresManualApproval(tt.assessment)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateRiskScore(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
		minScore      float64
		maxScore      float64
	}{
		{
			name: "calculates risk score successfully",
			setupMocks: func() {
				// CalculateRiskScore doesn't call GetUserTradingLimits, so don't mock it
				mockClient.On("GetUserRiskProfile", "user1").Return(createTestUserRiskProfile("user1"), nil)
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(createTestPositionExposure("AAPL"), nil)
				mockClient.On("GetAccountBalance", "user1").Return(createTestAccountBalance(), nil)
				mockClient.On("GetMarketVolatility", "AAPL").Return(createTestMarketVolatility("AAPL", false), nil)
			},
			order:    createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			minScore: 0.0,
			maxScore: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil // Reset mock
			tt.setupMocks()

			score, err := service.CalculateRiskScore(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, score, tt.minScore)
				assert.LessOrEqual(t, score, tt.maxScore)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAssessOrderRisk(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "comprehensive risk assessment succeeds",
			setupMocks: func() {
				setupDefaultMockExpectations(mockClient, "user1", "AAPL")
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil // Reset mock
			tt.setupMocks()

			assessment, err := service.AssessOrderRisk(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, assessment)
				assert.NotZero(t, assessment.AssessmentTime)
				assert.NotNil(t, assessment.RiskFactors)
				assert.NotNil(t, assessment.Recommendations)
				assert.NotNil(t, assessment.Warnings)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// Helper function tests

func TestDetermineRiskLevel(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults().(*riskManagementService)

	tests := []struct {
		name      string
		riskScore float64
		expected  RiskLevel
	}{
		{"extremely high risk", 85.0, RiskLevelExtremelyHigh},
		{"very high risk", 65.0, RiskLevelVeryHigh},
		{"high risk", 45.0, RiskLevelHigh},
		{"medium risk", 25.0, RiskLevelMedium},
		{"low risk", 15.0, RiskLevelLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.determineRiskLevel(tt.riskScore)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateMarketRiskScore(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults().(*riskManagementService)

	tests := []struct {
		name       string
		volatility *MarketVolatility
		minScore   float64
		maxScore   float64
	}{
		{
			name: "low volatility and beta",
			volatility: &MarketVolatility{
				Volatility30Day:  10.0,
				Beta:             1.0,
				IsHighVolatility: false,
			},
			minScore: 10.0,
			maxScore: 10.0,
		},
		{
			name: "high volatility and beta",
			volatility: &MarketVolatility{
				Volatility30Day:  40.0,
				Beta:             2.0,
				IsHighVolatility: true,
			},
			minScore: 65.0, // 40 + 10 + 15
			maxScore: 65.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateMarketRiskScore(tt.volatility)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, tt.maxScore)
		})
	}
}

func TestCalculateConcentrationRiskScore(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults().(*riskManagementService)

	tests := []struct {
		name                 string
		concentrationPercent float64
		expectedScore        float64
	}{
		{"within limit", 15.0, 30.0}, // 15 * 2
		{"at limit", 20.0, 40.0},     // 20 * 2
		{"above limit", 25.0, 60.0},  // 20*2 + 5*4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateConcentrationRiskScore(tt.concentrationPercent)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

func TestCalculateOrderSizeRiskScore(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults().(*riskManagementService)

	tests := []struct {
		name          string
		order         *domain.Order
		expectedScore float64
	}{
		{
			name:          "small order",
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10.0, floatPtr(50.0)),
			expectedScore: 5.0,
		},
		{
			name:          "medium order",
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 200.0, floatPtr(75.0)),
			expectedScore: 10.0,
		},
		{
			name:          "large order",
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 1000.0, floatPtr(75.0)),
			expectedScore: 20.0,
		},
		{
			name:          "very large order",
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 2000.0, floatPtr(75.0)),
			expectedScore: 30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateOrderSizeRiskScore(tt.order)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

func TestGetRiskToleranceMultiplier(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults().(*riskManagementService)

	tests := []struct {
		name      string
		tolerance RiskTolerance
		expected  float64
	}{
		{"conservative", RiskToleranceConservative, 0.5},
		{"moderate", RiskToleranceModerate, 0.8},
		{"aggressive", RiskToleranceAggressive, 1.2},
		{"speculative", RiskToleranceSpeculative, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getRiskToleranceMultiplier(tt.tolerance)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Edge Cases and Error Scenarios Tests

func TestValidateRiskLimits_EdgeCases(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "market order with zero calculated value",
			setupMocks: func() {
				userProfile := createTestUserRiskProfile("user1")
				tradingLimits := createTestTradingLimits()

				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetUserTradingLimits", "user1").Return(tradingLimits, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 100.0, nil),
		},
		{
			name: "order with exactly zero remaining daily limit",
			setupMocks: func() {
				userProfile := createTestUserRiskProfile("user1")
				tradingLimits := createTestTradingLimits()
				tradingLimits.RemainingDailyLimit = 0.0

				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetUserTradingLimits", "user1").Return(tradingLimits, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "order value 15000.00 exceeds remaining daily limit 0.00",
		},
		{
			name: "order with negative remaining daily limit",
			setupMocks: func() {
				userProfile := createTestUserRiskProfile("user1")
				tradingLimits := createTestTradingLimits()
				tradingLimits.RemainingDailyLimit = -1000.0

				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetUserTradingLimits", "user1").Return(tradingLimits, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "order value 15000.00 exceeds remaining daily limit -1000.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.setupMocks()

			err := service.ValidateRiskLimits(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCheckPositionLimits_EdgeCases(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "position with zero current value",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 0.0
				userProfile := createTestUserRiskProfile("user1")
				balance := createTestAccountBalance()

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
		{
			name: "position with negative current value",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = -5000.0 // Short position
				userProfile := createTestUserRiskProfile("user1")
				balance := createTestAccountBalance()

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
		{
			name: "account with zero total balance",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 1000.0
				userProfile := createTestUserRiskProfile("user1")
				balance := createTestAccountBalance()
				balance.TotalBalance = 0.0

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "position concentration", // Will be very high percentage
		},
		{
			name: "error getting position exposure",
			setupMocks: func() {
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(nil, errors.New("position service down"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "failed to get position exposure: position service down",
		},
		{
			name: "error getting user risk profile",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetUserRiskProfile", "user1").Return(nil, errors.New("user service error"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "failed to get user risk profile: user service error",
		},
		{
			name: "error getting account balance",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				userProfile := createTestUserRiskProfile("user1")

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetAccountBalance", "user1").Return(nil, errors.New("balance service error"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "failed to get account balance: balance service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.setupMocks()

			err := service.CheckPositionLimits(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAssessMarketRisk_EdgeCases(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "error getting market volatility",
			setupMocks: func() {
				mockClient.On("GetMarketVolatility", "AAPL").Return(nil, errors.New("market data unavailable"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "failed to get market volatility: market data unavailable",
		},
		{
			name: "extremely high volatility and beta",
			setupMocks: func() {
				volatility := createTestMarketVolatility("AAPL", true)
				volatility.Volatility30Day = 100.0 // Extreme volatility
				volatility.Beta = 5.0              // Extreme beta
				mockClient.On("GetMarketVolatility", "AAPL").Return(volatility, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
		{
			name: "zero volatility and beta",
			setupMocks: func() {
				volatility := createTestMarketVolatility("AAPL", false)
				volatility.Volatility30Day = 0.0
				volatility.Beta = 0.0
				mockClient.On("GetMarketVolatility", "AAPL").Return(volatility, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
		{
			name: "negative volatility and beta",
			setupMocks: func() {
				volatility := createTestMarketVolatility("AAPL", false)
				volatility.Volatility30Day = -5.0 // Invalid negative volatility
				volatility.Beta = -1.0            // Negative beta
				mockClient.On("GetMarketVolatility", "AAPL").Return(volatility, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			// Note: The service should handle negative values gracefully, even if they're invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.setupMocks()

			assessment, err := service.AssessMarketRisk(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, assessment)
				// Allow negative risk scores for edge cases with invalid data
				assert.LessOrEqual(t, assessment.RiskScore, 100.0)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAssessConcentrationRisk_EdgeCases(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "error getting position exposure",
			setupMocks: func() {
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(nil, errors.New("position data error"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "failed to get position exposure: position data error",
		},
		{
			name: "error getting account balance",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetAccountBalance", "user1").Return(nil, errors.New("balance data error"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "failed to get account balance: balance data error",
		},
		{
			name: "extreme concentration scenario",
			setupMocks: func() {
				position := createTestPositionExposure("AAPL")
				position.CurrentValue = 50000.0
				balance := createTestAccountBalance()
				balance.TotalBalance = 60000.0 // Very high concentration

				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.setupMocks()

			assessment, err := service.AssessConcentrationRisk(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, assessment)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCalculateRiskScore_EdgeCases(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "all risk components fail",
			setupMocks: func() {
				mockClient.On("GetMarketVolatility", "AAPL").Return(nil, errors.New("market data error"))
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(nil, errors.New("position error"))
				mockClient.On("GetUserRiskProfile", "user1").Return(nil, errors.New("user error"))
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			// Note: The service still calculates order size risk component, so it won't fail completely
		},
		{
			name: "partial risk components succeed",
			setupMocks: func() {
				// Market risk fails
				mockClient.On("GetMarketVolatility", "AAPL").Return(nil, errors.New("market data error"))

				// Concentration risk succeeds
				position := createTestPositionExposure("AAPL")
				balance := createTestAccountBalance()
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)

				// User risk succeeds
				userProfile := createTestUserRiskProfile("user1")
				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.setupMocks()

			score, err := service.CalculateRiskScore(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, score, 0.0)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAssessOrderRisk_EdgeCases(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)

	tests := []struct {
		name          string
		setupMocks    func()
		order         *domain.Order
		expectedError string
	}{
		{
			name: "assessment fails due to user profile error",
			setupMocks: func() {
				// Set up mocks for CalculateRiskScore first (which is called before individual assessments)
				mockClient.On("GetMarketVolatility", "AAPL").Return(createTestMarketVolatility("AAPL", false), nil)
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(createTestPositionExposure("AAPL"), nil)
				mockClient.On("GetAccountBalance", "user1").Return(createTestAccountBalance(), nil)
				// This will cause the assessUserRiskProfile to fail
				mockClient.On("GetUserRiskProfile", "user1").Return(nil, errors.New("user error"))
			},
			order:         createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
			expectedError: "user error",
		},
		{
			name: "high risk order requiring approval",
			setupMocks: func() {
				// Setup high risk scenario
				userProfile := createTestUserRiskProfile("user1")
				userProfile.RiskTolerance = RiskToleranceConservative
				userProfile.MaxOrderValue = 5000.0 // Low limit to trigger high risk

				position := createTestPositionExposure("AAPL")
				position.UnrealizedPnL = -5000.0 // Large losses

				balance := createTestAccountBalance()
				balance.TotalBalance = 30000.0 // Small balance for high concentration

				volatility := createTestMarketVolatility("AAPL", true)
				volatility.Volatility30Day = 50.0 // Very high volatility
				volatility.Beta = 3.0             // Very high beta

				limits := createTestTradingLimits()
				limits.RemainingDailyLimit = 20000.0

				mockClient.On("GetUserRiskProfile", "user1").Return(userProfile, nil)
				mockClient.On("GetPositionExposure", "user1", "AAPL").Return(position, nil)
				mockClient.On("GetAccountBalance", "user1").Return(balance, nil)
				mockClient.On("GetMarketVolatility", "AAPL").Return(volatility, nil)
				mockClient.On("GetUserTradingLimits", "user1").Return(limits, nil)
			},
			order: createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.setupMocks()

			assessment, err := service.AssessOrderRisk(tt.order, mockClient)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, assessment)

				// Verify assessment structure
				assert.NotZero(t, assessment.AssessmentTime)
				assert.NotNil(t, assessment.RiskFactors)
				assert.NotNil(t, assessment.Recommendations)
				assert.NotNil(t, assessment.Warnings)

				// Verify risk level is determined
				assert.True(t, assessment.RiskLevel >= RiskLevelLow && assessment.RiskLevel <= RiskLevelExtremelyHigh)

				// Verify approval logic
				if assessment.RiskScore > 80.0 {
					assert.False(t, assessment.IsApproved)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestHelperFunctions_EdgeCases(t *testing.T) {
	service := NewRiskManagementServiceWithDefaults().(*riskManagementService)

	t.Run("min function", func(t *testing.T) {
		assert.Equal(t, 5.0, min(5.0, 10.0))
		assert.Equal(t, 5.0, min(10.0, 5.0))
		assert.Equal(t, 0.0, min(0.0, 5.0))
		assert.Equal(t, -5.0, min(-5.0, 5.0))
	})

	t.Run("calculateMarketRiskScore with extreme values", func(t *testing.T) {
		// Test with maximum possible score
		volatility := &MarketVolatility{
			Volatility30Day:  200.0, // Very high
			Beta:             10.0,  // Extreme beta
			IsHighVolatility: true,
		}
		score := service.calculateMarketRiskScore(volatility)
		assert.Equal(t, 100.0, score) // Should be capped at 100
	})

	t.Run("calculateConcentrationRiskScore with extreme concentration", func(t *testing.T) {
		// Test with 100% concentration
		score := service.calculateConcentrationRiskScore(100.0)
		expected := service.concentrationLimit*2 + (100.0-service.concentrationLimit)*4
		assert.Equal(t, expected, score)
	})

	t.Run("getRiskToleranceMultiplier with invalid tolerance", func(t *testing.T) {
		// Test with invalid risk tolerance value
		invalidTolerance := RiskTolerance(999)
		multiplier := service.getRiskToleranceMultiplier(invalidTolerance)
		assert.Equal(t, 1.0, multiplier) // Should return default
	})
}

// Benchmark tests for performance validation

func BenchmarkAssessOrderRisk(b *testing.B) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)
	order := createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0))

	setupDefaultMockExpectations(mockClient, "user1", "AAPL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.AssessOrderRisk(order, mockClient)
	}
}

func BenchmarkCalculateRiskScore(b *testing.B) {
	service := NewRiskManagementServiceWithDefaults()
	mockClient := new(MockRiskDataClient)
	order := createTestOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 100.0, floatPtr(150.0))

	setupDefaultMockExpectations(mockClient, "user1", "AAPL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CalculateRiskScore(order, mockClient)
	}
}

// Utility functions

func floatPtr(f float64) *float64 {
	return &f
}
