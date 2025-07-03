package dto

// AssetDTO represents the database structure for assets
// This DTO handles the database mapping concerns
type AssetDTO struct {
	Symbol       string  `db:"symbol"`
	Quantity     float32 `db:"quantity"`
	AveragePrice float32 `db:"average_price"`
	LastPrice    float32 `db:"last_price"`
	Category     int     `db:"category"`
}
