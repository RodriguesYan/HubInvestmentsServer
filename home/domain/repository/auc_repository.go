package repository

import domain "HubInvestments/home/domain/model"

type AucRepository interface {
	GetPositionAggregation(userId string) ([]domain.AssetsModel, error)
}
