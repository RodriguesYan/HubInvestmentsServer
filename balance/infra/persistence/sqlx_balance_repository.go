package persistence

import (
	domain "HubInvestments/balance/domain/model"
	"HubInvestments/balance/domain/repository"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SQLXBalanceRepository struct {
	db *sqlx.DB
}

func NewSqlxBalanceRepository(db *sqlx.DB) repository.BalanceRepository {
	return &SQLXBalanceRepository{db: db}
}

func (r *SQLXBalanceRepository) GetBalance(userId string) (domain.BalanceModel, error) {
	balance, err := r.db.Queryx(
		`
		SELECT 	available_balance
		FROM	balances
		WHERE	user_id = $1
		`, userId)

	var balanceModel domain.BalanceModel

	if err != nil {
		println(err)
		return domain.BalanceModel{}, fmt.Errorf(err.Error())
	}

	for balance.Next() {
		var availableBalance float32

		if err := balance.Scan(&availableBalance); err != nil {
			return domain.BalanceModel{}, fmt.Errorf(err.Error())
		}

		balanceModel = domain.BalanceModel{
			AvailableBalance: availableBalance,
		}
	}

	return balanceModel, nil
}
