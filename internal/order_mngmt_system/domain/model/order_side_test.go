package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

func TestOrderSide_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    domain.OrderSide
		want bool
	}{
		{"Buy is valid", domain.OrderSideBuy, true},
		{"Sell is valid", domain.OrderSideSell, true},
		{"Invalid value", 0, false},
		{"Another invalid value", 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.IsValid())
		})
	}
}

func TestOrderSide_IsBuy(t *testing.T) {
	assert.True(t, domain.OrderSideBuy.IsBuy())
	assert.False(t, domain.OrderSideSell.IsBuy())
}

func TestOrderSide_IsSell(t *testing.T) {
	assert.True(t, domain.OrderSideSell.IsSell())
	assert.False(t, domain.OrderSideBuy.IsSell())
}

func TestOrderSide_RequiresPositionValidation(t *testing.T) {
	assert.True(t, domain.OrderSideSell.RequiresPositionValidation())
	assert.False(t, domain.OrderSideBuy.RequiresPositionValidation())
}

func TestOrderSide_GetDescription(t *testing.T) {
	tests := []struct {
		name string
		s    domain.OrderSide
		want string
	}{
		{"Buy description", domain.OrderSideBuy, "Buy order - purchasing assets"},
		{"Sell description", domain.OrderSideSell, "Sell order - selling assets from portfolio"},
		{"Unknown description", 0, "Unknown order side"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.GetDescription())
		})
	}
}
