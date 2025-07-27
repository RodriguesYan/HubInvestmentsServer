package persistence

import (
	repository "HubInvestments/internal/watchlist/domain/repository"
	"HubInvestments/shared/infra/database"
	"fmt"
	"strings"
)

type WatchlistRepository struct {
	db database.Database
}

func NewWatchlistRepository(db database.Database) repository.IWatchlistRepository {
	return &WatchlistRepository{db: db}
}

func (w *WatchlistRepository) GetWatchlist(userId string) ([]string, error) {
	var symbols string

	query := `SELECT symbols FROM watchlist WHERE user_id = $1`

	err := w.db.Get(&symbols, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist for user %s: %w", userId, err)
	}

	// Split comma-separated symbols into array
	symbolsArray := strings.Split(symbols, ",")

	return symbolsArray, nil
}
