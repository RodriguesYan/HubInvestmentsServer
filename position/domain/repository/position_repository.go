package repository

import domain "HubInvestments/position/domain/model"

type PositionRepository interface {
	GetPositionsByUserId(userId string) ([]domain.AssetsModel, error)
}
