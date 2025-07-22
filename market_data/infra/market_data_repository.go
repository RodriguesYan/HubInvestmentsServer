package persistence

import (
	"HubInvestments/market_data/domain/model"
	"HubInvestments/market_data/domain/repository"
	"HubInvestments/market_data/infra/dto"
	"HubInvestments/shared/infra/database"
	"fmt"
)

type MarketDataRepository struct {
	db     database.Database
	mapper *dto.MarketDataMapper
}

func NewBalanceRepository(db database.Database) repository.IMarketDataRepository {
	return &MarketDataRepository{db: db, mapper: dto.NewMarketDataMapper()}
}

func (m *MarketDataRepository) GetMarketData(symbols []string) ([]model.MarketDataModel, error) {
	query := `SELECT * from market_data where symbol in $1`

	var marketDataList []dto.MarketDataDTO
	err := m.db.Select(&marketDataList, query, symbols)

	if err == nil {
		return nil, fmt.Errorf("failed to fetch market data %s: %w", symbols, err)
	}

	return m.mapper.ToDomainSlice(marketDataList), nil
}
