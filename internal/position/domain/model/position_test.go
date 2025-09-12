package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPosition(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name         string
		userID       uuid.UUID
		symbol       string
		quantity     float64
		price        float64
		positionType PositionType
		wantError    bool
		errorMsg     string
	}{
		{
			name:         "Valid position",
			userID:       userID,
			symbol:       "AAPL",
			quantity:     100.0,
			price:        150.0,
			positionType: PositionTypeLong,
			wantError:    false,
		},
		{
			name:         "Invalid user ID",
			userID:       uuid.Nil,
			symbol:       "AAPL",
			quantity:     100.0,
			price:        150.0,
			positionType: PositionTypeLong,
			wantError:    true,
			errorMsg:     "user ID cannot be empty",
		},
		{
			name:         "Empty symbol",
			userID:       userID,
			symbol:       "",
			quantity:     100.0,
			price:        150.0,
			positionType: PositionTypeLong,
			wantError:    true,
			errorMsg:     "symbol cannot be empty",
		},
		{
			name:         "Zero quantity",
			userID:       userID,
			symbol:       "AAPL",
			quantity:     0.0,
			price:        150.0,
			positionType: PositionTypeLong,
			wantError:    true,
			errorMsg:     "quantity must be greater than zero",
		},
		{
			name:         "Negative quantity",
			userID:       userID,
			symbol:       "AAPL",
			quantity:     -100.0,
			price:        150.0,
			positionType: PositionTypeLong,
			wantError:    true,
			errorMsg:     "quantity must be greater than zero",
		},
		{
			name:         "Zero price",
			userID:       userID,
			symbol:       "AAPL",
			quantity:     100.0,
			price:        0.0,
			positionType: PositionTypeLong,
			wantError:    true,
			errorMsg:     "price must be greater than zero",
		},
		{
			name:         "Negative price",
			userID:       userID,
			symbol:       "AAPL",
			quantity:     100.0,
			price:        -150.0,
			positionType: PositionTypeLong,
			wantError:    true,
			errorMsg:     "price must be greater than zero",
		},
		{
			name:         "Invalid position type",
			userID:       userID,
			symbol:       "AAPL",
			quantity:     100.0,
			price:        150.0,
			positionType: PositionType("INVALID"),
			wantError:    true,
			errorMsg:     "invalid position type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position, err := NewPosition(tt.userID, tt.symbol, tt.quantity, tt.price, tt.positionType)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewPosition() expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("NewPosition() error = %v, want %v", err.Error(), tt.errorMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewPosition() unexpected error: %v", err)
				return
			}

			if position.UserID != tt.userID {
				t.Errorf("NewPosition().UserID = %v, want %v", position.UserID, tt.userID)
			}

			if position.Symbol != tt.symbol {
				t.Errorf("NewPosition().Symbol = %v, want %v", position.Symbol, tt.symbol)
			}

			if position.Quantity != tt.quantity {
				t.Errorf("NewPosition().Quantity = %v, want %v", position.Quantity, tt.quantity)
			}

			if position.AveragePrice != tt.price {
				t.Errorf("NewPosition().AveragePrice = %v, want %v", position.AveragePrice, tt.price)
			}

			expectedInvestment := tt.quantity * tt.price
			if position.TotalInvestment != expectedInvestment {
				t.Errorf("NewPosition().TotalInvestment = %v, want %v", position.TotalInvestment, expectedInvestment)
			}

			if position.Status != PositionStatusActive {
				t.Errorf("NewPosition().Status = %v, want %v", position.Status, PositionStatusActive)
			}
		})
	}
}

func TestPosition_CalculateNewAveragePrice(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	tests := []struct {
		name        string
		newQuantity float64
		newPrice    float64
		expected    float64
		wantError   bool
	}{
		{
			name:        "Valid calculation",
			newQuantity: 50.0,
			newPrice:    160.0,
			expected:    153.333333, // (100*150 + 50*160) / (100+50) = 23000/150 = 153.33
			wantError:   false,
		},
		{
			name:        "Equal prices",
			newQuantity: 100.0,
			newPrice:    150.0,
			expected:    150.0,
			wantError:   false,
		},
		{
			name:        "Higher price",
			newQuantity: 100.0,
			newPrice:    200.0,
			expected:    175.0, // (100*150 + 100*200) / (100+100) = 35000/200 = 175
			wantError:   false,
		},
		{
			name:        "Zero quantity",
			newQuantity: 0.0,
			newPrice:    160.0,
			expected:    0,
			wantError:   true,
		},
		{
			name:        "Zero price",
			newQuantity: 50.0,
			newPrice:    0.0,
			expected:    0,
			wantError:   true,
		},
		{
			name:        "Negative quantity",
			newQuantity: -50.0,
			newPrice:    160.0,
			expected:    0,
			wantError:   true,
		},
		{
			name:        "Negative price",
			newQuantity: 50.0,
			newPrice:    -160.0,
			expected:    0,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := position.CalculateNewAveragePrice(tt.newQuantity, tt.newPrice)

			if tt.wantError {
				if err == nil {
					t.Errorf("CalculateNewAveragePrice() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateNewAveragePrice() unexpected error: %v", err)
				return
			}

			// Use a tolerance for floating point comparison
			tolerance := 0.01
			if got < tt.expected-tolerance || got > tt.expected+tolerance {
				t.Errorf("CalculateNewAveragePrice() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPosition_CanSell(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	tests := []struct {
		name         string
		sellQuantity float64
		expected     bool
	}{
		{"Can sell less than holding", 50.0, true},
		{"Can sell exactly holding", 100.0, true},
		{"Cannot sell more than holding", 150.0, false},
		{"Cannot sell zero", 0.0, false},
		{"Cannot sell negative", -50.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := position.CanSell(tt.sellQuantity); got != tt.expected {
				t.Errorf("Position.CanSell() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPosition_UpdateQuantity_BuyOrder(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	// Test buying more shares
	err := position.UpdateQuantity(50.0, 160.0, true)
	if err != nil {
		t.Errorf("UpdateQuantity() unexpected error: %v", err)
		return
	}

	// Check quantity increased
	expectedQuantity := 150.0
	if position.Quantity != expectedQuantity {
		t.Errorf("UpdateQuantity() quantity = %v, want %v", position.Quantity, expectedQuantity)
	}

	// Check average price recalculated: (100*150 + 50*160) / 150 = 23000/150 = 153.33
	expectedAvgPrice := 153.333333
	tolerance := 0.01
	if position.AveragePrice < expectedAvgPrice-tolerance || position.AveragePrice > expectedAvgPrice+tolerance {
		t.Errorf("UpdateQuantity() average price = %v, want %v", position.AveragePrice, expectedAvgPrice)
	}

	// Check total investment updated
	expectedInvestment := 150.0 * position.AveragePrice
	if position.TotalInvestment < expectedInvestment-tolerance || position.TotalInvestment > expectedInvestment+tolerance {
		t.Errorf("UpdateQuantity() total investment = %v, want %v", position.TotalInvestment, expectedInvestment)
	}
}

func TestPosition_UpdateQuantity_SellOrder(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	// Test selling some shares
	err := position.UpdateQuantity(30.0, 160.0, false)
	if err != nil {
		t.Errorf("UpdateQuantity() unexpected error: %v", err)
		return
	}

	// Check quantity decreased
	expectedQuantity := 70.0
	if position.Quantity != expectedQuantity {
		t.Errorf("UpdateQuantity() quantity = %v, want %v", position.Quantity, expectedQuantity)
	}

	// Check average price unchanged for sell orders
	expectedAvgPrice := 150.0
	if position.AveragePrice != expectedAvgPrice {
		t.Errorf("UpdateQuantity() average price = %v, want %v (should not change on sell)", position.AveragePrice, expectedAvgPrice)
	}

	// Check status changed to partial
	if position.Status != PositionStatusPartial {
		t.Errorf("UpdateQuantity() status = %v, want %v", position.Status, PositionStatusPartial)
	}
}

func TestPosition_UpdateQuantity_SellAllShares(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	// Test selling all shares
	err := position.UpdateQuantity(100.0, 160.0, false)
	if err != nil {
		t.Errorf("UpdateQuantity() unexpected error: %v", err)
		return
	}

	// Check quantity is zero
	if position.Quantity != 0.0 {
		t.Errorf("UpdateQuantity() quantity = %v, want 0", position.Quantity)
	}

	// Check status changed to closed
	if position.Status != PositionStatusClosed {
		t.Errorf("UpdateQuantity() status = %v, want %v", position.Status, PositionStatusClosed)
	}
}

func TestPosition_UpdateQuantity_InvalidOperations(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	tests := []struct {
		name          string
		tradeQuantity float64
		tradePrice    float64
		isBuyOrder    bool
		expectedError string
	}{
		{
			name:          "Zero quantity",
			tradeQuantity: 0.0,
			tradePrice:    160.0,
			isBuyOrder:    true,
			expectedError: "trade quantity must be greater than zero",
		},
		{
			name:          "Zero price",
			tradeQuantity: 50.0,
			tradePrice:    0.0,
			isBuyOrder:    true,
			expectedError: "trade price must be greater than zero",
		},
		{
			name:          "Insufficient quantity to sell",
			tradeQuantity: 150.0,
			tradePrice:    160.0,
			isBuyOrder:    false,
			expectedError: "insufficient quantity to sell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := position.UpdateQuantity(tt.tradeQuantity, tt.tradePrice, tt.isBuyOrder)
			if err == nil {
				t.Errorf("UpdateQuantity() expected error but got none")
				return
			}

			if err.Error()[:len(tt.expectedError)] != tt.expectedError {
				t.Errorf("UpdateQuantity() error = %v, want error containing %v", err.Error(), tt.expectedError)
			}
		})
	}
}

func TestPosition_UpdateCurrentPrice(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	// Test updating current price
	currentPrice := 165.0
	err := position.UpdateCurrentPrice(currentPrice)
	if err != nil {
		t.Errorf("UpdateCurrentPrice() unexpected error: %v", err)
		return
	}

	// Check current price set
	if position.CurrentPrice != currentPrice {
		t.Errorf("UpdateCurrentPrice() current price = %v, want %v", position.CurrentPrice, currentPrice)
	}

	// Check market value calculated
	expectedMarketValue := 100.0 * 165.0 // quantity * current price
	if position.MarketValue != expectedMarketValue {
		t.Errorf("UpdateCurrentPrice() market value = %v, want %v", position.MarketValue, expectedMarketValue)
	}

	// Check unrealized PnL calculated
	expectedPnL := expectedMarketValue - position.TotalInvestment // 16500 - 15000 = 1500
	if position.UnrealizedPnL != expectedPnL {
		t.Errorf("UpdateCurrentPrice() unrealized PnL = %v, want %v", position.UnrealizedPnL, expectedPnL)
	}

	// Check unrealized PnL percentage: (1500 / 15000) * 100 = 10%
	expectedPnLPct := 10.0
	if position.UnrealizedPnLPct != expectedPnLPct {
		t.Errorf("UpdateCurrentPrice() unrealized PnL%% = %v, want %v", position.UnrealizedPnLPct, expectedPnLPct)
	}

	// Test with invalid price
	err = position.UpdateCurrentPrice(0.0)
	if err == nil {
		t.Errorf("UpdateCurrentPrice() expected error for zero price")
	}
}

func TestPosition_Validate(t *testing.T) {
	userID := uuid.New()

	// Valid position
	validPosition, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)
	if err := validPosition.Validate(); err != nil {
		t.Errorf("Validate() unexpected error for valid position: %v", err)
	}

	// Test various invalid scenarios
	tests := []struct {
		name     string
		modify   func(*Position)
		errorMsg string
	}{
		{
			name: "Empty ID",
			modify: func(p *Position) {
				p.ID = uuid.Nil
			},
			errorMsg: "position ID cannot be empty",
		},
		{
			name: "Empty User ID",
			modify: func(p *Position) {
				p.UserID = uuid.Nil
			},
			errorMsg: "user ID cannot be empty",
		},
		{
			name: "Empty Symbol",
			modify: func(p *Position) {
				p.Symbol = ""
			},
			errorMsg: "symbol cannot be empty",
		},
		{
			name: "Negative Quantity",
			modify: func(p *Position) {
				p.Quantity = -10.0
			},
			errorMsg: "quantity cannot be negative",
		},
		{
			name: "Closed status with quantity",
			modify: func(p *Position) {
				p.Status = PositionStatusClosed
				p.Quantity = 10.0
			},
			errorMsg: "closed position cannot have quantity greater than zero",
		},
		{
			name: "Active status with zero quantity",
			modify: func(p *Position) {
				p.Status = PositionStatusActive
				p.Quantity = 0.0
			},
			errorMsg: "active position must have quantity greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)
			tt.modify(position)

			err := position.Validate()
			if err == nil {
				t.Errorf("Validate() expected error but got none")
				return
			}

			if err.Error() != tt.errorMsg {
				t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestPosition_UtilityMethods(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	// Test CanBeClosed
	if !position.CanBeClosed() {
		t.Errorf("CanBeClosed() = false, want true for active position with quantity")
	}

	// Test IsEmpty
	if position.IsEmpty() {
		t.Errorf("IsEmpty() = true, want false for position with quantity")
	}

	// Test GetRealizedValue
	currentPrice := 160.0
	expectedValue := 100.0 * 160.0
	if got := position.GetRealizedValue(currentPrice); got != expectedValue {
		t.Errorf("GetRealizedValue() = %v, want %v", got, expectedValue)
	}

	// Test String method
	str := position.String()
	if str == "" {
		t.Errorf("String() returned empty string")
	}

	// Test with closed position
	position.Status = PositionStatusClosed
	position.Quantity = 0.0

	if position.CanBeClosed() {
		t.Errorf("CanBeClosed() = true, want false for closed position")
	}

	if !position.IsEmpty() {
		t.Errorf("IsEmpty() = false, want true for position with zero quantity")
	}
}
