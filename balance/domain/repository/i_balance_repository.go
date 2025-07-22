package repository

import domain "HubInvestments/balance/domain/model"

type IBalanceRepository interface {
	GetBalance(userId string) (domain.BalanceModel, error)
}
