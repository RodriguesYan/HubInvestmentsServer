package persistence

import (
	"HubInvestments/internal/market_data/domain/model"
	"HubInvestments/internal/market_data/domain/repository"
	"HubInvestments/internal/market_data/infra/dto"
	"HubInvestments/shared/infra/database"
	"fmt"
	"strings"
)

type MarketDataRepository struct {
	db     database.Database
	mapper *dto.MarketDataMapper
}

func NewMarketDataRepository(db database.Database) repository.IMarketDataRepository {
	return &MarketDataRepository{db: db, mapper: dto.NewMarketDataMapper()}
}

func (m *MarketDataRepository) GetMarketData(symbols []string) ([]model.MarketDataModel, error) {
	// Create placeholders for the IN clause
	placeholders := make([]string, len(symbols))
	args := make([]interface{}, len(symbols))

	for i, symbol := range symbols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = symbol
	}

	query := fmt.Sprintf("SELECT * FROM market_data WHERE symbol IN (%s)",
		strings.Join(placeholders, ","))

	var marketDataList []dto.MarketDataDTO
	err := m.db.Select(&marketDataList, query, args...)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data %v: %w", symbols, err)
	}

	return m.mapper.ToDomainSlice(marketDataList), nil
}
