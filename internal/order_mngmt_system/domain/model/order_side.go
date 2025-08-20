package domain

import (
	"encoding/json"
	"fmt"
)

// OrderSide represents the side of an order (buy or sell)
// @Description Order side enumeration
type OrderSide int32

const (
	// OrderSideBuy represents a buy order
	OrderSideBuy OrderSide = 1

	// OrderSideSell represents a sell order
	OrderSideSell OrderSide = 2
)

// IsValid checks if the order side is valid
func (s OrderSide) IsValid() bool {
	switch s {
	case OrderSideBuy, OrderSideSell:
		return true
	default:
		return false
	}
}

// IsBuy checks if this is a buy order
func (s OrderSide) IsBuy() bool {
	return s == OrderSideBuy
}

// IsSell checks if this is a sell order
func (s OrderSide) IsSell() bool {
	return s == OrderSideSell
}

// RequiresPositionValidation checks if the order side requires position validation
func (s OrderSide) RequiresPositionValidation() bool {
	return s == OrderSideSell
}

// ParseOrderSide parses a string into an OrderSide
func ParseOrderSide(s string) (OrderSide, error) {
	switch s {
	case "BUY":
		return OrderSideBuy, nil
	case "SELL":
		return OrderSideSell, nil
	default:
		return 0, fmt.Errorf("invalid order side: %s", s)
	}
}

// String returns the string representation of the order side
func (s OrderSide) String() string {
	switch s {
	case OrderSideBuy:
		return "BUY"
	case OrderSideSell:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON implements json.Marshaler interface
func (s OrderSide) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler interface
func (s *OrderSide) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parsed, err := ParseOrderSide(str)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}

func (s OrderSide) GetDescription() string {
	switch s {
	case OrderSideBuy:
		return "Buy order - purchasing assets"
	case OrderSideSell:
		return "Sell order - selling assets from portfolio"
	default:
		return "Unknown order side"
	}
}
