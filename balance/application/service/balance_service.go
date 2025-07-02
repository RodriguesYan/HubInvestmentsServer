package service

import (
	domain "HubInvestments/balance/domain/model"
	"HubInvestments/balance/domain/repository"
)

type BalanceServiceInterface interface {
	GetBalance(userId string) (domain.BalanceModel, error)
}

type BalanceService struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

func (s *BalanceService) GetBalance(userId string) (domain.BalanceModel, error) {
	return s.repo.GetBalance(userId)
}
