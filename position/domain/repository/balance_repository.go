package repository

import domain "HubInvestments/position/domain/model"

type BalanceRepository interface {
	GetBalance(userId string) (domain.BalanceModel, error)
}
