package usecase

import (
	domain "HubInvestments/balance/domain/model"
	"HubInvestments/balance/domain/repository"
)

type GetBalanceUseCase struct {
	repo repository.BalanceRepository
}

func NewGetBalanceUseCase(repo repository.BalanceRepository) *GetBalanceUseCase {
	return &GetBalanceUseCase{repo: repo}
}

func (uc *GetBalanceUseCase) Execute(userId string) (domain.BalanceModel, error) {
	balance, err := uc.repo.GetBalance(userId)
	if err != nil {
		return domain.BalanceModel{}, err
	}

	return balance, nil
}
