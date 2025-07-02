package domain

type BalanceModel struct {
	AvailableBalance float32 `json:"availableBalance" db:"available_balance"`
}
