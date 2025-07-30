package repository

import domain "HubInvestments/internal/position/domain/model"

type PositionRepository interface {
	GetPositionsByUserId(userId string) ([]domain.AssetModel, error)
}
