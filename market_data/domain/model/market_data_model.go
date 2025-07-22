package model

type MarketDataModel struct {
	Symbol    string
	Name      string
	LastQuote float32
	Category  int //TODO: criar enum pra esse cara
}
