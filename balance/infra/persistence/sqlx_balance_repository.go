package persistence

import (
	domain "HubInvestments/balance/domain/model"
	"HubInvestments/balance/domain/repository"
	"database/sql"
	"errors"
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
	var availableBalance float32
	query := `SELECT available_balance FROM balances WHERE user_id = $1`

	err := r.db.Get(&availableBalance, query, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no rows are found, it's not a fatal error.
			// It means the user has no balance record, which we treat as a balance of 0.
			return domain.BalanceModel{}, nil
		}
		// For any other error, we wrap it and return.
		return domain.BalanceModel{}, fmt.Errorf("failed to get balance for user %s: %w", userId, err)
	}

	return domain.BalanceModel{AvailableBalance: availableBalance}, nil
}
