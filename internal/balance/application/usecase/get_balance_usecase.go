package usecase

import (
	domain "HubInvestments/internal/balance/domain/model"
	repository "HubInvestments/internal/balance/domain/repository"
)

type GetBalanceUseCase struct {
	repo repository.IBalanceRepository
}

func NewGetBalanceUseCase(repo repository.IBalanceRepository) *GetBalanceUseCase {
	return &GetBalanceUseCase{repo: repo}
}

func (uc *GetBalanceUseCase) Execute(userId string) (domain.BalanceModel, error) {
	balance, err := uc.repo.GetBalance(userId)
	if err != nil {
		return domain.BalanceModel{}, err
	}

	return balance, nil
}
