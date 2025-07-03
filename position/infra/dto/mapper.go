package dto

import domain "HubInvestments/position/domain/model"

// AssetMapper handles conversion between AssetDTO and domain.AssetsModel
type AssetMapper struct{}

// NewAssetMapper creates a new asset mapper
func NewAssetMapper() *AssetMapper {
	return &AssetMapper{}
}

// ToDomain converts AssetDTO to domain.AssetsModel
func (m *AssetMapper) ToDomain(dto AssetDTO) domain.AssetsModel {
	return domain.AssetsModel{
		Symbol:       dto.Symbol,
		Quantity:     dto.Quantity,
		AveragePrice: dto.AveragePrice,
		LastPrice:    dto.LastPrice,
		Category:     dto.Category,
	}
}

// ToDTO converts domain.AssetsModel to AssetDTO
func (m *AssetMapper) ToDTO(model domain.AssetsModel) AssetDTO {
	return AssetDTO{
		Symbol:       model.Symbol,
		Quantity:     model.Quantity,
		AveragePrice: model.AveragePrice,
		LastPrice:    model.LastPrice,
		Category:     model.Category,
	}
}

// ToDomainSlice converts a slice of AssetDTO to slice of domain.AssetsModel
func (m *AssetMapper) ToDomainSlice(dtos []AssetDTO) []domain.AssetsModel {
	models := make([]domain.AssetsModel, len(dtos))
	for i, dto := range dtos {
		models[i] = m.ToDomain(dto)
	}
	return models
}
