package persistence

import (
	domain "HubInvestments/position/domain/model"
	"HubInvestments/position/domain/repository"
	"HubInvestments/position/infra/dto"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SQLXAucRepository struct {
	db     *sqlx.DB
	mapper *dto.AssetMapper
}

func NewSQLXAucRepository(db *sqlx.DB) repository.AucRepository {
	return &SQLXAucRepository{
		db:     db,
		mapper: dto.NewAssetMapper(),
	}
}

func (r *SQLXAucRepository) GetPositionAggregation(userId string) ([]domain.AssetsModel, error) {
	query := `
	SELECT 	i.symbol, 
			p.average_price, 
			p.quantity, 
			i.category, 
			i.last_price
	FROM positions p 
	join instruments i on p.instrument_id = i.id 
	where p.user_id = $1`

	var assetDTOs []dto.AssetDTO
	err := r.db.Select(&assetDTOs, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get position aggregation: %w", err)
	}

	// Convert DTOs to domain models using mapper
	return r.mapper.ToDomainSlice(assetDTOs), nil
}
