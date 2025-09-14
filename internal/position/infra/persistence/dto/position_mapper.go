package dto

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	domain "HubInvestments/internal/position/domain/model"
)

type PositionMapper struct{}

func NewPositionMapper() *PositionMapper {
	return &PositionMapper{}
}

func (m *PositionMapper) ToDomain(dto *PositionDTO) (*domain.Position, error) {
	return dto.ToDomain()
}

func (m *PositionMapper) ToDomainList(dtos []*PositionDTO) ([]*domain.Position, error) {
	positions := make([]*domain.Position, len(dtos))
	for i, dto := range dtos {
		position, err := dto.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert DTO to domain model at index %d: %w", i, err)
		}
		positions[i] = position
	}
	return positions, nil
}

func (m *PositionMapper) CreateDTOForInsert(position *domain.Position) (*PositionDTO, error) {
	if position == nil {
		return nil, errors.New("domain position cannot be nil")
	}

	dto := &PositionDTO{
		ID:              position.ID,
		UserID:          position.UserID,
		Symbol:          position.Symbol,
		Quantity:        position.Quantity,
		AveragePrice:    position.AveragePrice,
		TotalInvestment: position.TotalInvestment,
		PositionType:    position.PositionType.String(),
		Status:          position.Status.String(),
		CreatedAt:       position.CreatedAt,
		UpdatedAt:       position.UpdatedAt,
	}

	if position.CurrentPrice != 0 {
		dto.CurrentPrice = sql.NullFloat64{Float64: position.CurrentPrice, Valid: true}
	}
	if position.MarketValue != 0 {
		dto.MarketValue = sql.NullFloat64{Float64: position.MarketValue, Valid: true}
	}
	if position.UnrealizedPnL != 0 {
		dto.UnrealizedPnL = sql.NullFloat64{Float64: position.UnrealizedPnL, Valid: true}
	}
	if position.UnrealizedPnLPct != 0 {
		dto.UnrealizedPnLPct = sql.NullFloat64{Float64: position.UnrealizedPnLPct, Valid: true}
	}
	if position.LastTradeAt != nil {
		dto.LastTradeAt = sql.NullTime{Time: *position.LastTradeAt, Valid: true}
	}

	return dto, nil
}

func (m *PositionMapper) CreateDTOForUpdate(position *domain.Position) (*PositionDTO, error) {
	if position == nil {
		return nil, errors.New("domain position cannot be nil")
	}

	dto := &PositionDTO{
		ID:              position.ID,
		Quantity:        position.Quantity,
		AveragePrice:    position.AveragePrice,
		TotalInvestment: position.TotalInvestment,
		Status:          position.Status.String(),
		UpdatedAt:       time.Now(),
	}

	if position.CurrentPrice != 0 {
		dto.CurrentPrice = sql.NullFloat64{Float64: position.CurrentPrice, Valid: true}
	}
	if position.MarketValue != 0 {
		dto.MarketValue = sql.NullFloat64{Float64: position.MarketValue, Valid: true}
	}
	if position.UnrealizedPnL != 0 {
		dto.UnrealizedPnL = sql.NullFloat64{Float64: position.UnrealizedPnL, Valid: true}
	}
	if position.UnrealizedPnLPct != 0 {
		dto.UnrealizedPnLPct = sql.NullFloat64{Float64: position.UnrealizedPnLPct, Valid: true}
	}
	if position.LastTradeAt != nil {
		dto.LastTradeAt = sql.NullTime{Time: *position.LastTradeAt, Valid: true}
	}

	return dto, nil
}
