package dto

import (
	"fmt"

	domain "HubInvestments/internal/position/domain/model"

	"github.com/google/uuid"
)

// PositionMapper handles conversion between Position domain model and PositionDTO
type PositionMapper struct{}

// NewPositionMapper creates a new instance of PositionMapper
func NewPositionMapper() *PositionMapper {
	return &PositionMapper{}
}

// ToDTO converts a Position domain model to PositionDTO for database persistence
func (m *PositionMapper) ToDTO(position *domain.Position) (*PositionDTO, error) {
	if position == nil {
		return nil, fmt.Errorf("position cannot be nil: %w", ErrInvalidDomainModel)
	}

	dto := &PositionDTO{
		ID:               position.ID,
		UserID:           position.UserID,
		Symbol:           position.Symbol,
		Quantity:         position.Quantity,
		AveragePrice:     position.AveragePrice,
		TotalInvestment:  position.TotalInvestment,
		CurrentPrice:     position.CurrentPrice,
		MarketValue:      position.MarketValue,
		UnrealizedPnL:    position.UnrealizedPnL,
		UnrealizedPnLPct: position.UnrealizedPnLPct,
		PositionType:     string(position.PositionType),
		Status:           string(position.Status),
		CreatedAt:        position.CreatedAt,
		UpdatedAt:        position.UpdatedAt,
		LastTradeAt:      position.LastTradeAt,
	}

	if err := dto.Validate(); err != nil {
		return nil, fmt.Errorf("DTO validation failed: %w", err)
	}

	return dto, nil
}

// ToDomain converts a PositionDTO to Position domain model
func (m *PositionMapper) ToDomain(dto *PositionDTO) (*domain.Position, error) {
	if dto == nil {
		return nil, fmt.Errorf("DTO cannot be nil: %w", ErrInvalidDomainModel)
	}

	if err := dto.Validate(); err != nil {
		return nil, fmt.Errorf("DTO validation failed: %w", err)
	}

	// Convert string value objects to domain types
	positionType, err := domain.NewPositionType(dto.PositionType)
	if err != nil {
		return nil, fmt.Errorf("invalid position type '%s': %w", dto.PositionType, err)
	}

	positionStatus, err := domain.NewPositionStatus(dto.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid position status '%s': %w", dto.Status, err)
	}

	// Create domain position without triggering business logic validation
	// This is important when loading from database as the data should already be valid
	position := &domain.Position{
		ID:               dto.ID,
		UserID:           dto.UserID,
		Symbol:           dto.Symbol,
		Quantity:         dto.Quantity,
		AveragePrice:     dto.AveragePrice,
		TotalInvestment:  dto.TotalInvestment,
		CurrentPrice:     dto.CurrentPrice,
		MarketValue:      dto.MarketValue,
		UnrealizedPnL:    dto.UnrealizedPnL,
		UnrealizedPnLPct: dto.UnrealizedPnLPct,
		PositionType:     positionType,
		Status:           positionStatus,
		CreatedAt:        dto.CreatedAt,
		UpdatedAt:        dto.UpdatedAt,
		LastTradeAt:      dto.LastTradeAt,
	}

	// Clear any domain events since this is loaded from persistence
	position.ClearEvents()

	return position, nil
}

// ToDTOList converts a slice of Position domain models to PositionDTOs
func (m *PositionMapper) ToDTOList(positions []*domain.Position) ([]*PositionDTO, error) {
	if positions == nil {
		return nil, nil
	}

	dtos := make([]*PositionDTO, len(positions))
	for i, position := range positions {
		dto, err := m.ToDTO(position)
		if err != nil {
			return nil, fmt.Errorf("failed to convert position %d: %w", i, err)
		}
		dtos[i] = dto
	}

	return dtos, nil
}

// ToDomainList converts a slice of PositionDTOs to Position domain models
func (m *PositionMapper) ToDomainList(dtos []*PositionDTO) ([]*domain.Position, error) {
	if dtos == nil {
		return nil, nil
	}

	positions := make([]*domain.Position, len(dtos))
	for i, dto := range dtos {
		position, err := m.ToDomain(dto)
		if err != nil {
			return nil, fmt.Errorf("failed to convert DTO %d: %w", i, err)
		}
		positions[i] = position
	}

	return positions, nil
}

// CreateDTOFromDomainForInsert creates a DTO optimized for database insertion
// This includes generating new UUIDs if needed and setting proper timestamps
func (m *PositionMapper) CreateDTOForInsert(position *domain.Position) (*PositionDTO, error) {
	dto, err := m.ToDTO(position)
	if err != nil {
		return nil, err
	}

	// Ensure we have a valid UUID for new records
	if dto.ID == uuid.Nil {
		dto.ID = uuid.New()
	}

	// Don't override domain model timestamps as they contain business logic
	// The domain model should manage its own temporal state

	return dto, nil
}

// CreateDTOForUpdate creates a DTO optimized for database updates
// This preserves the ID and creation timestamp while updating modification time
func (m *PositionMapper) CreateDTOForUpdate(position *domain.Position) (*PositionDTO, error) {
	dto, err := m.ToDTO(position)
	if err != nil {
		return nil, err
	}

	// Validate that we have an ID for updates
	if dto.ID == uuid.Nil {
		return nil, fmt.Errorf("position ID required for update: %w", ErrInvalidPositionID)
	}

	return dto, nil
}

// MapPositionForQuery creates a DTO with query parameters for database filtering
type PositionQueryDTO struct {
	UserID       *uuid.UUID `db:"user_id,omitempty"`
	Symbol       *string    `db:"symbol,omitempty"`
	Status       *string    `db:"status,omitempty"`
	PositionType *string    `db:"position_type,omitempty"`
	Limit        *int       `db:"limit,omitempty"`
	Offset       *int       `db:"offset,omitempty"`
}

// CreateQueryDTO creates a DTO for database query filtering
func (m *PositionMapper) CreateQueryDTO() *PositionQueryDTO {
	return &PositionQueryDTO{}
}

// WithUserID adds user ID filter to query
func (q *PositionQueryDTO) WithUserID(userID uuid.UUID) *PositionQueryDTO {
	q.UserID = &userID
	return q
}

// WithSymbol adds symbol filter to query
func (q *PositionQueryDTO) WithSymbol(symbol string) *PositionQueryDTO {
	q.Symbol = &symbol
	return q
}

// WithStatus adds status filter to query
func (q *PositionQueryDTO) WithStatus(status string) *PositionQueryDTO {
	q.Status = &status
	return q
}

// WithPositionType adds position type filter to query
func (q *PositionQueryDTO) WithPositionType(positionType string) *PositionQueryDTO {
	q.PositionType = &positionType
	return q
}

// WithPagination adds pagination parameters to query
func (q *PositionQueryDTO) WithPagination(limit, offset int) *PositionQueryDTO {
	q.Limit = &limit
	q.Offset = &offset
	return q
}
