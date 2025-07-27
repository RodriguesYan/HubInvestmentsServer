package repository

type IWatchlistRepository interface {
	GetWatchlist(userId string) ([]string, error)
}
