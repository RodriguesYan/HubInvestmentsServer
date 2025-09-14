package domain

import (
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestPositionSplitter_SplitByQuantity(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)
	position.CurrentPrice = 155.0
	position.MarketValue = position.Quantity * position.CurrentPrice
	position.UnrealizedPnL = position.MarketValue - position.TotalInvestment

	splitter := NewPositionSplitter()

	tests := []struct {
		name           string
		quantities     []float64
		expectError    bool
		expectedSplits int
	}{
		{
			name:           "Valid split into two positions",
			quantities:     []float64{60.0, 40.0},
			expectError:    false,
			expectedSplits: 2,
		},
		{
			name:           "Valid split into three positions",
			quantities:     []float64{30.0, 30.0, 30.0},
			expectError:    false,
			expectedSplits: 3,
		},
		{
			name:        "Split exceeds total quantity",
			quantities:  []float64{60.0, 60.0},
			expectError: true,
		},
		{
			name:        "Empty quantities array",
			quantities:  []float64{},
			expectError: true,
		},
		{
			name:           "Contains zero quantity",
			quantities:     []float64{50.0, 0.0, 50.0},
			expectError:    false,
			expectedSplits: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := splitter.SplitByQuantity(position, tt.quantities)

			if tt.expectError {
				if err == nil {
					t.Errorf("SplitByQuantity() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("SplitByQuantity() unexpected error: %v", err)
				return
			}

			if len(result.SplitPositions) != tt.expectedSplits {
				t.Errorf("SplitByQuantity() got %d splits, want %d", len(result.SplitPositions), tt.expectedSplits)
			}

			// Verify each split position
			for i, split := range result.SplitPositions {
				if split.UserID != position.UserID {
					t.Errorf("Split position %d UserID = %v, want %v", i, split.UserID, position.UserID)
				}
				if split.Symbol != position.Symbol {
					t.Errorf("Split position %d Symbol = %v, want %v", i, split.Symbol, position.Symbol)
				}
				if split.AveragePrice != position.AveragePrice {
					t.Errorf("Split position %d AveragePrice = %v, want %v", i, split.AveragePrice, position.AveragePrice)
				}
			}
		})
	}
}

func TestPositionSplitter_SplitByPercentage(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)

	splitter := NewPositionSplitter()

	tests := []struct {
		name           string
		percentages    []float64
		expectError    bool
		expectedSplits int
	}{
		{
			name:           "Valid 60-40 split",
			percentages:    []float64{60.0, 40.0},
			expectError:    false,
			expectedSplits: 2,
		},
		{
			name:           "Valid 33-33-33 split (total 99%)",
			percentages:    []float64{33.0, 33.0, 33.0},
			expectError:    false,
			expectedSplits: 3,
		},
		{
			name:        "Percentages exceed 100%",
			percentages: []float64{60.0, 50.0},
			expectError: true,
		},
		{
			name:        "Invalid negative percentage",
			percentages: []float64{60.0, -10.0},
			expectError: true,
		},
		{
			name:        "Invalid zero percentage",
			percentages: []float64{60.0, 0.0},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := splitter.SplitByPercentage(position, tt.percentages)

			if tt.expectError {
				if err == nil {
					t.Errorf("SplitByPercentage() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("SplitByPercentage() unexpected error: %v", err)
				return
			}

			if len(result.SplitPositions) != tt.expectedSplits {
				t.Errorf("SplitByPercentage() got %d splits, want %d", len(result.SplitPositions), tt.expectedSplits)
			}
		})
	}
}

func TestRoundToDecimalPlaces(t *testing.T) {
	tests := []struct {
		name          string
		value         float64
		decimalPlaces int
		expected      float64
	}{
		{"Round to 2 decimal places", 123.456789, 2, 123.46},
		{"Round to 4 decimal places", 123.456789, 4, 123.4568},
		{"No rounding needed", 123.45, 2, 123.45},
		{"Round zero", 0.0, 2, 0.0},
		{"Round negative", -123.456, 2, -123.46},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToDecimalPlaces(tt.value, tt.decimalPlaces)
			if result != tt.expected {
				t.Errorf("RoundToDecimalPlaces() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsFinanciallyEqual(t *testing.T) {
	tests := []struct {
		name      string
		a         float64
		b         float64
		precision int
		expected  bool
	}{
		{"Equal values", 123.45, 123.45, 2, true},
		{"Within tolerance", 123.451, 123.452, 2, true},
		{"Outside tolerance", 123.45, 123.47, 2, false},
		{"High precision equal", 123.123456, 123.123456, 6, true},
		{"High precision different", 123.123456, 123.123466, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFinanciallyEqual(tt.a, tt.b, tt.precision)
			if result != tt.expected {
				t.Errorf("IsFinanciallyEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateFinancialValue(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		fieldName   string
		minValue    float64
		maxValue    float64
		precision   int
		expectError bool
	}{
		{"Valid value", 100.0, "price", 0.0, 1000.0, 2, false},
		{"Value too low", -10.0, "price", 0.0, 1000.0, 2, true},
		{"Value too high", 2000.0, "price", 0.0, 1000.0, 2, true},
		{"NaN value", math.NaN(), "price", 0.0, 1000.0, 2, true},
		{"Infinite value", math.Inf(1), "price", 0.0, 1000.0, 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFinancialValue(tt.value, tt.fieldName, tt.minValue, tt.maxValue, tt.precision)

			if tt.expectError && err == nil {
				t.Errorf("ValidateFinancialValue() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("ValidateFinancialValue() unexpected error: %v", err)
			}
		})
	}
}

func TestPositionBusinessValidator_ValidateTradeOperation(t *testing.T) {
	userID := uuid.New()
	position, _ := NewPosition(userID, "AAPL", 100.0, 150.0, PositionTypeLong)
	validator := &PositionBusinessValidator{}

	tests := []struct {
		name           string
		tradeQuantity  float64
		tradePrice     float64
		isBuyOrder     bool
		expectedErrors int
	}{
		{
			name:           "Valid buy order",
			tradeQuantity:  50.0,
			tradePrice:     160.0,
			isBuyOrder:     true,
			expectedErrors: 0,
		},
		{
			name:           "Valid sell order",
			tradeQuantity:  50.0,
			tradePrice:     160.0,
			isBuyOrder:     false,
			expectedErrors: 0,
		},
		{
			name:           "Trade value too low",
			tradeQuantity:  0.001,
			tradePrice:     0.001,
			isBuyOrder:     true,
			expectedErrors: 1,
		},
		{
			name:           "Sell more than owned",
			tradeQuantity:  150.0,
			tradePrice:     160.0,
			isBuyOrder:     false,
			expectedErrors: 1,
		},
		{
			name:           "Buy exceeds maximum quantity",
			tradeQuantity:  1e10,
			tradePrice:     160.0,
			isBuyOrder:     true,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateTradeOperation(position, tt.tradeQuantity, tt.tradePrice, tt.isBuyOrder)

			if len(errors) != tt.expectedErrors {
				t.Errorf("ValidateTradeOperation() got %d errors, want %d", len(errors), tt.expectedErrors)
				for _, err := range errors {
					t.Logf("Error: %v", err)
				}
			}
		})
	}
}

func TestPositionMerger_MergePositions(t *testing.T) {
	userID := uuid.New()
	merger := &PositionMerger{}

	// Create multiple positions for the same user and symbol
	position1, _ := NewPosition(userID, "AAPL", 50.0, 140.0, PositionTypeLong)
	position2, _ := NewPosition(userID, "AAPL", 30.0, 160.0, PositionTypeLong)
	position3, _ := NewPosition(userID, "AAPL", 20.0, 180.0, PositionTypeLong)

	tests := []struct {
		name        string
		positions   []*Position
		expectError bool
	}{
		{
			name:        "Merge two positions",
			positions:   []*Position{position1, position2},
			expectError: false,
		},
		{
			name:        "Merge three positions",
			positions:   []*Position{position1, position2, position3},
			expectError: false,
		},
		{
			name:        "Single position (no merge needed)",
			positions:   []*Position{position1},
			expectError: false,
		},
		{
			name:        "Empty positions array",
			positions:   []*Position{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, err := merger.MergePositions(tt.positions)

			if tt.expectError {
				if err == nil {
					t.Errorf("MergePositions() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("MergePositions() unexpected error: %v", err)
				return
			}

			if merged == nil {
				t.Errorf("MergePositions() returned nil position")
				return
			}

			// For multiple positions, verify merged metrics
			if len(tt.positions) > 1 {
				expectedQuantity := 0.0
				expectedInvestment := 0.0
				for _, pos := range tt.positions {
					expectedQuantity += pos.Quantity
					expectedInvestment += pos.TotalInvestment
				}

				tolerance := 0.01
				if math.Abs(merged.Quantity-expectedQuantity) > tolerance {
					t.Errorf("MergePositions() quantity = %v, want %v", merged.Quantity, expectedQuantity)
				}

				if math.Abs(merged.TotalInvestment-expectedInvestment) > tolerance {
					t.Errorf("MergePositions() total investment = %v, want %v", merged.TotalInvestment, expectedInvestment)
				}

				expectedAvgPrice := expectedInvestment / expectedQuantity
				if math.Abs(merged.AveragePrice-expectedAvgPrice) > tolerance {
					t.Errorf("MergePositions() average price = %v, want %v", merged.AveragePrice, expectedAvgPrice)
				}
			}
		})
	}
}

func TestPositionMerger_MergePositions_ValidationErrors(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	merger := &PositionMerger{}

	position1, _ := NewPosition(userID1, "AAPL", 50.0, 140.0, PositionTypeLong)
	position2, _ := NewPosition(userID2, "AAPL", 30.0, 160.0, PositionTypeLong)  // Different user
	position3, _ := NewPosition(userID1, "GOOGL", 20.0, 180.0, PositionTypeLong) // Different symbol
	position4, _ := NewPosition(userID1, "AAPL", 25.0, 170.0, PositionTypeShort) // Different type

	tests := []struct {
		name        string
		positions   []*Position
		expectedErr string
	}{
		{
			name:        "Different users",
			positions:   []*Position{position1, position2},
			expectedErr: "cannot merge positions for different users",
		},
		{
			name:        "Different symbols",
			positions:   []*Position{position1, position3},
			expectedErr: "cannot merge positions for different symbols",
		},
		{
			name:        "Different position types",
			positions:   []*Position{position1, position4},
			expectedErr: "cannot merge positions of different types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := merger.MergePositions(tt.positions)

			if err == nil {
				t.Errorf("MergePositions() expected error but got none")
				return
			}

			if err.Error() != tt.expectedErr {
				t.Errorf("MergePositions() error = %v, want %v", err.Error(), tt.expectedErr)
			}
		})
	}
}
