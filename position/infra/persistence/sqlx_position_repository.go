package persistence

import (
	domain "HubInvestments/position/domain/model"
	"HubInvestments/position/domain/repository"
	"HubInvestments/position/infra/dto"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SQLXPositionRepository struct {
	db     *sqlx.DB
	mapper *dto.AssetMapper
}

func NewSQLXPositionRepository(db *sqlx.DB) repository.PositionRepository {
	return &SQLXPositionRepository{
		db:     db,
		mapper: dto.NewAssetMapper(),
	}
}

func (r *SQLXPositionRepository) GetPositionsByUserId(userId string) ([]domain.AssetsModel, error) {
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
