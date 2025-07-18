package service

import (
	domain "HubInvestments/position/domain/model"
	"HubInvestments/position/domain/repository"
)

type AucServiceInterface interface {
	GetAucAggregation(userId string) ([]domain.AssetsModel, error)
}

type AucService struct {
	repo repository.PositionRepository
}

func NewAucService(repo repository.PositionRepository) *AucService {
	return &AucService{repo: repo}
}

func (s *AucService) GetAucAggregation(userId string) ([]domain.AssetsModel, error) {
	return s.repo.GetPositionsByUserId(userId)
}
