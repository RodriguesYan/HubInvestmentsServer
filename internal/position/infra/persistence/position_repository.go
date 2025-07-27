package persistence

import (
	domain "HubInvestments/internal/position/domain/model"
	repository "HubInvestments/internal/position/domain/repository"
	dto "HubInvestments/internal/position/infra/dto"
	"HubInvestments/shared/infra/database"
	"fmt"
)

// PositionRepository implements the repository interface using the database abstraction
type PositionRepository struct {
	db     database.Database
	mapper *dto.AssetMapper
}

// NewPositionRepository creates a new position repository using the database abstraction
func NewPositionRepository(db database.Database) repository.PositionRepository {
	return &PositionRepository{
		db:     db,
		mapper: dto.NewAssetMapper(),
	}
}

func (r *PositionRepository) GetPositionsByUserId(userId string) ([]domain.AssetsModel, error) {
	query := `
	SELECT 	i.symbol, 
			p.average_price, 
			p.quantity, 
			i.category, 
			i.last_price
	FROM positions p 
	JOIN instruments i ON p.instrument_id = i.id 
	WHERE p.user_id = $1`

	var assetDTOs []dto.AssetDTO
	err := r.db.Select(&assetDTOs, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions for user %s: %w", userId, err)
	}

	// Convert DTOs to domain models using mapper
	return r.mapper.ToDomainSlice(assetDTOs), nil
}
