package dto

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
)

// PositionDTO represents the database structure for positions_v2 table
type PositionDTO struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	UserID           uuid.UUID  `db:"user_id" json:"userId"`
	Symbol           string     `db:"symbol" json:"symbol"`
	Quantity         float64    `db:"quantity" json:"quantity"`
	AveragePrice     float64    `db:"average_price" json:"averagePrice"`
	TotalInvestment  float64    `db:"total_investment" json:"totalInvestment"`
	CurrentPrice     float64    `db:"current_price" json:"currentPrice"`
	MarketValue      float64    `db:"market_value" json:"marketValue"`
	UnrealizedPnL    float64    `db:"unrealized_pnl" json:"unrealizedPnL"`
	UnrealizedPnLPct float64    `db:"unrealized_pnl_pct" json:"unrealizedPnLPct"`
	PositionType     string     `db:"position_type" json:"positionType"`
	Status           string     `db:"status" json:"status"`
	CreatedAt        time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updatedAt"`
	LastTradeAt      *time.Time `db:"last_trade_at" json:"lastTradeAt,omitempty"`
}

// Ensure PositionDTO implements driver.Valuer and sql.Scanner if needed
var _ driver.Valuer = (*PositionDTO)(nil)

// Value implements driver.Valuer interface for custom serialization if needed
func (p PositionDTO) Value() (driver.Value, error) {
	return p, nil
}

// TableName returns the qualified table name for this DTO
func (PositionDTO) TableName() string {
	return "yanrodrigues.positions_v2"
}

// NewPositionDTO creates a new PositionDTO with default values
func NewPositionDTO() *PositionDTO {
	return &PositionDTO{
		ID:               uuid.New(),
		CurrentPrice:     0,
		MarketValue:      0,
		UnrealizedPnL:    0,
		UnrealizedPnLPct: 0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// IsEmpty returns true if this is an uninitialized DTO
func (p *PositionDTO) IsEmpty() bool {
	return p.ID == uuid.Nil
}

// HasCurrentPrice returns true if current price data is available
func (p *PositionDTO) HasCurrentPrice() bool {
	return p.CurrentPrice > 0
}

// CalculateMarketValue calculates market value based on quantity and current price
func (p *PositionDTO) CalculateMarketValue() float64 {
	if !p.HasCurrentPrice() {
		return 0
	}
	return p.Quantity * p.CurrentPrice
}

// CalculateUnrealizedPnL calculates unrealized P&L based on market value and investment
func (p *PositionDTO) CalculateUnrealizedPnL() float64 {
	if !p.HasCurrentPrice() {
		return 0
	}
	return p.CalculateMarketValue() - p.TotalInvestment
}

// CalculateUnrealizedPnLPct calculates unrealized P&L percentage
func (p *PositionDTO) CalculateUnrealizedPnLPct() float64 {
	if !p.HasCurrentPrice() || p.TotalInvestment <= 0 {
		return 0
	}
	return (p.CalculateUnrealizedPnL() / p.TotalInvestment) * 100
}

// UpdateMarketData updates all market-related fields consistently
func (p *PositionDTO) UpdateMarketData(currentPrice float64) {
	p.CurrentPrice = currentPrice
	p.MarketValue = p.CalculateMarketValue()
	p.UnrealizedPnL = p.CalculateUnrealizedPnL()
	p.UnrealizedPnLPct = p.CalculateUnrealizedPnLPct()
	p.UpdatedAt = time.Now()
}

// Validate performs basic DTO validation
func (p *PositionDTO) Validate() error {
	if p.ID == uuid.Nil {
		return ErrInvalidPositionID
	}
	if p.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if p.Symbol == "" {
		return ErrInvalidSymbol
	}
	if p.Quantity < 0 {
		return ErrNegativeQuantity
	}
	if p.AveragePrice < 0 {
		return ErrNegativePrice
	}
	if p.TotalInvestment < 0 {
		return ErrNegativeInvestment
	}
	if p.PositionType != "LONG" && p.PositionType != "SHORT" {
		return ErrInvalidPositionType
	}
	if p.Status != "ACTIVE" && p.Status != "PARTIAL" && p.Status != "CLOSED" {
		return ErrInvalidStatus
	}
	return nil
}
