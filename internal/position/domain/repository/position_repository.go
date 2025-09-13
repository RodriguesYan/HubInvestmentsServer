package repository

import (
	domain "HubInvestments/internal/position/domain/model"
	"context"

	"github.com/google/uuid"
)

// IPositionRepository defines the interface for position persistence operations
type IPositionRepository interface {
	// Legacy method for compatibility
	GetPositionsByUserId(userId string) ([]domain.AssetModel, error)

	// New Position domain model methods
	FindByID(ctx context.Context, positionID uuid.UUID) (*domain.Position, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error)
	FindByUserIDAndSymbol(ctx context.Context, userID uuid.UUID, symbol string) (*domain.Position, error)
	FindActivePositions(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error)
	Save(ctx context.Context, position *domain.Position) error
	Update(ctx context.Context, position *domain.Position) error
	Delete(ctx context.Context, positionID uuid.UUID) error

	// Query methods
	ExistsForUser(ctx context.Context, userID uuid.UUID, symbol string) (bool, error)
	CountPositionsForUser(ctx context.Context, userID uuid.UUID) (int, error)
	GetTotalInvestmentForUser(ctx context.Context, userID uuid.UUID) (float64, error)
}

// For backward compatibility
type PositionRepository = IPositionRepository
