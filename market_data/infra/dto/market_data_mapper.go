package dto

import (
	"HubInvestments/market_data/domain/model"
)

// AssetMapper handles conversion between AssetDTO and domain.AssetsModel
type MarketDataMapper struct{}

// NewAssetMapper creates a new asset mapper
func NewMarketDataMapper() *MarketDataMapper {
	return &MarketDataMapper{}
}

// ToDomain converts AssetDTO to domain.AssetsModel
func (m *MarketDataMapper) ToDomain(dto MarketDataDTO) model.MarketDataModel {
	return model.MarketDataModel{
		Symbol:    dto.Symbol,
		Category:  dto.Category,
		LastQuote: dto.LastQuote,
		Name:      dto.Name,
	}
}

// ToDTO converts domain.AssetsModel to AssetDTO
func (m *MarketDataMapper) ToDTO(model model.MarketDataModel) MarketDataDTO {
	return MarketDataDTO{
		Symbol:    model.Symbol,
		Category:  model.Category,
		Name:      model.Name,
		LastQuote: model.LastQuote,
	}
}

// ToDomainSlice converts a slice of AssetDTO to slice of domain.AssetsModel
func (m *MarketDataMapper) ToDomainSlice(dtos []MarketDataDTO) []model.MarketDataModel {
	models := make([]model.MarketDataModel, len(dtos))
	for i, dto := range dtos {
		models[i] = m.ToDomain(dto)
	}
	return models
}
