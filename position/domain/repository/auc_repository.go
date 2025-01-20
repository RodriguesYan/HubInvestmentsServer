package repository

import domain "HubInvestments/position/domain/model"

type AucRepository interface {
	GetPositionAggregation(userId string) ([]domain.AssetsModel, error)
}
