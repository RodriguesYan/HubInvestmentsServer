package service

import (
	domain "HubInvestments/home/domain/model"
	"HubInvestments/home/domain/repository"
)

type AucServiceInterface interface {
	GetAucAggregation(userId string) ([]domain.AssetsModel, error)
}

type AucService struct {
	repo repository.AucRepository
}

func NewAucService(repo repository.AucRepository) *AucService {
	return &AucService{repo: repo}
}

func (s *AucService) GetAucAggregation(userId string) ([]domain.AssetsModel, error) {
	return s.repo.GetPositionAggregation(userId)
}
