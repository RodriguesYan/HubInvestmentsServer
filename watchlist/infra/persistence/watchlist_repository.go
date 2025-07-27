package persistence

import (
	"HubInvestments/shared/infra/database"
	"HubInvestments/watchlist/domain/repository"
)

type WatchlistRepository struct {
	db database.Database
}

func NewWatchlistRepository(db database.Database) repository.IWatchlistRepository {
	return &WatchlistRepository{db: db}
}

func (w *WatchlistRepository) GetWatchlist(userId string) ([]string, error) {
	result := w.db.Get()
}
