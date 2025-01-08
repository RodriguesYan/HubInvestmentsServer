package service

import (
	domain "HubInvestments/home/domain/model"
	"HubInvestments/home/domain/repository"
)

type AucService struct {
	repo repository.AucRepository
}

func NewUserService(repo repository.AucRepository) *AucService {
	return &AucService{repo: repo}
}

func (s *AucService) GetAucAggregation(userId string) ([]domain.AssetsModel, error) {
	return s.repo.GetPositionAggregation(userId)
}
