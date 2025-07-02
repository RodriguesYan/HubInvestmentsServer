package repository

import domain "HubInvestments/balance/domain/model"

type BalanceRepository interface {
	GetBalance(userId string) (domain.BalanceModel, error)
}
