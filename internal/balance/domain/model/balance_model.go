package domain

// BalanceModel represents the user's balance information
// @Description User balance information
type BalanceModel struct {
	AvailableBalance float32 `json:"availableBalance" db:"available_balance" example:"15000.50"`
}
