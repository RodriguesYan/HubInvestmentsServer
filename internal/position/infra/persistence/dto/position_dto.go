package dto

import (
	"database/sql"
	"fmt"
	"time"

	domain "HubInvestments/internal/position/domain/model"

	"github.com/google/uuid"
)

// PositionDTO represents the data transfer object for a Position in the database.
type PositionDTO struct {
	ID               uuid.UUID       `db:"id"`
	UserID           uuid.UUID       `db:"user_id"`
	Symbol           string          `db:"symbol"`
	Quantity         float64         `db:"quantity"`
	AveragePrice     float64         `db:"average_price"`
	TotalInvestment  float64         `db:"total_investment"`
	CurrentPrice     sql.NullFloat64 `db:"current_price"`
	MarketValue      sql.NullFloat64 `db:"market_value"`
	UnrealizedPnL    sql.NullFloat64 `db:"unrealized_pnl"`
	UnrealizedPnLPct sql.NullFloat64 `db:"unrealized_pnl_pct"`
	PositionType     string          `db:"position_type"`
	Status           string          `db:"status"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
	LastTradeAt      sql.NullTime    `db:"last_trade_at"`
}

// ToDomain converts a PositionDTO to a domain.Position model.
func (dto *PositionDTO) ToDomain() (*domain.Position, error) {
	positionType, err := domain.NewPositionType(dto.PositionType)
	if err != nil {
		return nil, fmt.Errorf("invalid position type in DTO: %w", err)
	}
	positionStatus, err := domain.NewPositionStatus(dto.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid position status in DTO: %w", err)
	}

	position, err := domain.NewPosition(
		dto.UserID,
		dto.Symbol,
		dto.Quantity,
		dto.AveragePrice,
		positionType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain position from DTO: %w", err)
	}

	position.ID = dto.ID
	position.TotalInvestment = dto.TotalInvestment
	position.CreatedAt = dto.CreatedAt
	position.UpdatedAt = dto.UpdatedAt
	position.Status = positionStatus

	if dto.CurrentPrice.Valid {
		position.CurrentPrice = dto.CurrentPrice.Float64
	}
	if dto.MarketValue.Valid {
		position.MarketValue = dto.MarketValue.Float64
	}
	if dto.UnrealizedPnL.Valid {
		position.UnrealizedPnL = dto.UnrealizedPnL.Float64
	}
	if dto.UnrealizedPnLPct.Valid {
		position.UnrealizedPnLPct = dto.UnrealizedPnLPct.Float64
	}
	if dto.LastTradeAt.Valid {
		position.LastTradeAt = &dto.LastTradeAt.Time
	}

	return position, nil
}
