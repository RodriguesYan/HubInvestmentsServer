package persistence

import (
	domain "HubInvestments/internal/balance/domain/model"
	repository "HubInvestments/internal/balance/domain/repository"
	"HubInvestments/shared/infra/database"
	"database/sql"
	"errors"
	"fmt"
)

// BalanceRepository implements the repository interface using the database abstraction
type BalanceRepository struct {
	db database.Database
}

// NewBalanceRepository creates a new balance repository using the database abstraction
func NewBalanceRepository(db database.Database) repository.IBalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) GetBalance(userId string) (domain.BalanceModel, error) {
	var availableBalance float32
	query := `SELECT available_balance FROM balances WHERE user_id = $1`

	err := r.db.Get(&availableBalance, query, userId)

	if err == nil {
		return domain.BalanceModel{AvailableBalance: availableBalance}, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		// If no rows are found, it's not a fatal error.
		// It means the user has no balance record, which we treat as a balance of 0.
		return domain.BalanceModel{}, nil
	}
	// For any other error, we wrap it and return.
	return domain.BalanceModel{}, fmt.Errorf("failed to get balance for user %s: %w", userId, err)

}
