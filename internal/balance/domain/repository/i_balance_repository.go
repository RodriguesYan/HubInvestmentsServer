package repository

import domain "HubInvestments/internal/balance/domain/model"

type IBalanceRepository interface {
	GetBalance(userId string) (domain.BalanceModel, error)
}
