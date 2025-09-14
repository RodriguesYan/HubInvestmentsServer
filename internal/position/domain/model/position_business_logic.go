package domain

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// Business logic utilities for Position management
// These utilities support advanced position operations like splitting, precision handling, and complex validations

const (
	DefaultPricePrecision    = 8    // 8 decimal places for prices
	DefaultQuantityPrecision = 6    // 6 decimal places for quantities
	MinimumTradeValue        = 0.01 // Minimum $0.01 trade value
	MaxPositionQuantity      = 1e9  // Maximum position quantity
)

type PositionSplitter struct {
	precision int
}

func NewPositionSplitter() *PositionSplitter {
	return &PositionSplitter{precision: DefaultQuantityPrecision}
}

type SplitResult struct {
	OriginalPosition *Position
	SplitPositions   []*Position
	TotalQuantity    float64
	ValidationErrors []string
}

// Splits a position into multiple smaller positions with specified quantities
// This is useful for partial position management or portfolio diversification
func (ps *PositionSplitter) SplitByQuantity(position *Position, quantities []float64) (*SplitResult, error) {
	if position == nil {
		return nil, errors.New("position cannot be nil")
	}

	if len(quantities) == 0 {
		return nil, errors.New("quantities array cannot be empty")
	}

	result := &SplitResult{
		OriginalPosition: position,
		SplitPositions:   make([]*Position, 0),
		ValidationErrors: make([]string, 0),
	}

	// Validate total quantities don't exceed position quantity
	totalSplitQuantity := 0.0
	for _, qty := range quantities {
		if qty <= 0 {
			result.ValidationErrors = append(result.ValidationErrors,
				"all split quantities must be positive")
			continue
		}
		totalSplitQuantity += qty
	}

	result.TotalQuantity = totalSplitQuantity

	if totalSplitQuantity > position.Quantity {
		return result, errors.New("total split quantities exceed position quantity")
	}

	// Create split positions
	for i, qty := range quantities {
		if qty <= 0 {
			continue // Skip invalid quantities
		}

		splitPosition := &Position{
			ID:              uuid.New(),
			UserID:          position.UserID,
			Symbol:          position.Symbol,
			Quantity:        RoundToDecimalPlaces(qty, ps.precision),
			AveragePrice:    position.AveragePrice,
			TotalInvestment: RoundToDecimalPlaces(qty*position.AveragePrice, DefaultPricePrecision),
			CurrentPrice:    position.CurrentPrice,
			MarketValue:     RoundToDecimalPlaces(qty*position.CurrentPrice, DefaultPricePrecision),
			PositionType:    position.PositionType,
			Status:          PositionStatusActive,
			CreatedAt:       position.CreatedAt,
			UpdatedAt:       position.UpdatedAt,
			LastTradeAt:     position.LastTradeAt,
			events:          make([]DomainEvent, 0),
		}

		// Calculate P&L for split position
		if splitPosition.TotalInvestment > 0 {
			splitPosition.UnrealizedPnL = splitPosition.MarketValue - splitPosition.TotalInvestment
			splitPosition.UnrealizedPnLPct = (splitPosition.UnrealizedPnL / splitPosition.TotalInvestment) * 100
		}

		// Validate split position
		if err := splitPosition.Validate(); err != nil {
			result.ValidationErrors = append(result.ValidationErrors,
				fmt.Sprintf("split position %d validation failed: %s", i+1, err.Error()))
			continue
		}

		result.SplitPositions = append(result.SplitPositions, splitPosition)
	}

	return result, nil
}

func (ps *PositionSplitter) SplitByPercentage(position *Position, percentages []float64) (*SplitResult, error) {
	if position == nil {
		return nil, errors.New("position cannot be nil")
	}

	// Validate percentages sum to <= 100%
	totalPercentage := 0.0
	for _, pct := range percentages {
		if pct <= 0 || pct > 100 {
			return nil, errors.New("all percentages must be between 0 and 100")
		}
		totalPercentage += pct
	}

	if totalPercentage > 100 {
		return nil, errors.New("total percentages cannot exceed 100%")
	}

	// Convert percentages to quantities
	quantities := make([]float64, len(percentages))
	for i, pct := range percentages {
		quantities[i] = position.Quantity * (pct / 100.0)
	}

	return ps.SplitByQuantity(position, quantities)
}

// Provides utilities for handling financial precision
type PrecisionHelper struct{}

func RoundToDecimalPlaces(value float64, decimalPlaces int) float64 {
	multiplier := math.Pow(10, float64(decimalPlaces))
	return math.Round(value*multiplier) / multiplier
}

// Compares two float64 values with financial precision tolerance
func IsFinanciallyEqual(a, b float64, precision int) bool {
	tolerance := math.Pow(10, float64(-precision))
	return math.Abs(a-b) < tolerance
}

func ValidateFinancialValue(value float64, fieldName string, minValue, maxValue float64, precision int) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return errors.New(fieldName + " must be a valid finite number")
	}

	if value < minValue {
		return errors.New(fieldName + " must be at least " + formatFloat(minValue, precision))
	}

	if value > maxValue {
		return errors.New(fieldName + " cannot exceed " + formatFloat(maxValue, precision))
	}

	return nil
}

func formatFloat(value float64, precision int) string {
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, value)
}

type PositionBusinessValidator struct{}

func (v *PositionBusinessValidator) ValidateTradeOperation(position *Position, tradeQuantity, tradePrice float64, isBuyOrder bool) []ValidationError {
	var errors []ValidationError

	// Validate trade value meets minimum threshold
	tradeValue := tradeQuantity * tradePrice
	if tradeValue < MinimumTradeValue {
		errors = append(errors, ValidationError{
			Field:   "tradeValue",
			Message: "trade value must be at least $0.01",
			Value:   tradeValue,
		})
	}

	if !isValidPrecision(tradeQuantity, DefaultQuantityPrecision) {
		errors = append(errors, ValidationError{
			Field:   "tradeQuantity",
			Message: "quantity precision exceeds maximum allowed decimal places",
			Value:   tradeQuantity,
		})
	}

	if !isValidPrecision(tradePrice, DefaultPricePrecision) {
		errors = append(errors, ValidationError{
			Field:   "tradePrice",
			Message: "price precision exceeds maximum allowed decimal places",
			Value:   tradePrice,
		})
	}

	// Validate sell operations don't create negative positions
	if !isBuyOrder {
		if tradeQuantity > position.Quantity {
			errors = append(errors, ValidationError{
				Field:   "sellQuantity",
				Message: "cannot sell more shares than owned",
				Value:   tradeQuantity,
			})
		}
	}

	// Validate position won't exceed maximum quantity after buy
	if isBuyOrder {
		newQuantity := position.Quantity + tradeQuantity
		if newQuantity > MaxPositionQuantity {
			errors = append(errors, ValidationError{
				Field:   "newQuantity",
				Message: "position quantity would exceed maximum allowed",
				Value:   newQuantity,
			})
		}
	}

	return errors
}

type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value"`
}

func (ve ValidationError) Error() string {
	return ve.Field + ": " + ve.Message
}

// Checks if a float64 value has acceptable decimal precision
func isValidPrecision(value float64, maxDecimalPlaces int) bool {
	multiplier := math.Pow(10, float64(maxDecimalPlaces))
	return value*multiplier == math.Floor(value*multiplier+0.5)
}

type PositionMerger struct{}

// Combines multiple positions of the same symbol for the same user
// This is useful for consolidating positions that were split or acquired separately
func (pm *PositionMerger) MergePositions(positions []*Position) (*Position, error) {
	if len(positions) == 0 {
		return nil, errors.New("no positions to merge")
	}

	if len(positions) == 1 {
		return positions[0], nil
	}

	// Validate all positions can be merged
	basePosition := positions[0]
	for i, pos := range positions {
		if pos.UserID != basePosition.UserID {
			return nil, errors.New("cannot merge positions for different users")
		}
		if pos.Symbol != basePosition.Symbol {
			return nil, errors.New("cannot merge positions for different symbols")
		}
		if pos.PositionType != basePosition.PositionType {
			return nil, errors.New("cannot merge positions of different types")
		}
		if !pos.Status.CanBeUpdated() {
			return nil, fmt.Errorf("position %d cannot be updated", i)
		}
	}

	// Calculate merged position metrics
	totalQuantity := 0.0
	totalInvestment := 0.0
	earliestCreatedAt := basePosition.CreatedAt
	latestUpdatedAt := basePosition.UpdatedAt
	var latestTradeAt *time.Time

	for _, pos := range positions {
		totalQuantity += pos.Quantity
		totalInvestment += pos.TotalInvestment

		if pos.CreatedAt.Before(earliestCreatedAt) {
			earliestCreatedAt = pos.CreatedAt
		}
		if pos.UpdatedAt.After(latestUpdatedAt) {
			latestUpdatedAt = pos.UpdatedAt
		}
		if pos.LastTradeAt != nil {
			if latestTradeAt == nil || pos.LastTradeAt.After(*latestTradeAt) {
				latestTradeAt = pos.LastTradeAt
			}
		}
	}

	// Calculate weighted average price
	var averagePrice float64
	if totalQuantity > 0 {
		averagePrice = totalInvestment / totalQuantity
	}

	mergedPosition := &Position{
		ID:              uuid.New(),
		UserID:          basePosition.UserID,
		Symbol:          basePosition.Symbol,
		Quantity:        RoundToDecimalPlaces(totalQuantity, DefaultQuantityPrecision),
		AveragePrice:    RoundToDecimalPlaces(averagePrice, DefaultPricePrecision),
		TotalInvestment: RoundToDecimalPlaces(totalInvestment, DefaultPricePrecision),
		CurrentPrice:    basePosition.CurrentPrice, // Use most recent current price
		PositionType:    basePosition.PositionType,
		Status:          PositionStatusActive,
		CreatedAt:       earliestCreatedAt,
		UpdatedAt:       latestUpdatedAt,
		LastTradeAt:     latestTradeAt,
		events:          make([]DomainEvent, 0),
	}

	// Recalculate market value and P&L if current price is available
	if mergedPosition.CurrentPrice > 0 {
		mergedPosition.MarketValue = RoundToDecimalPlaces(
			mergedPosition.Quantity*mergedPosition.CurrentPrice, DefaultPricePrecision)
		mergedPosition.UnrealizedPnL = mergedPosition.MarketValue - mergedPosition.TotalInvestment
		if mergedPosition.TotalInvestment > 0 {
			mergedPosition.UnrealizedPnLPct = (mergedPosition.UnrealizedPnL / mergedPosition.TotalInvestment) * 100
		}
	}

	if err := mergedPosition.Validate(); err != nil {
		return nil, errors.New("merged position validation failed: " + err.Error())
	}

	return mergedPosition, nil
}
