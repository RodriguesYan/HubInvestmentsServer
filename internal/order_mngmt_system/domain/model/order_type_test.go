package domain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

func TestAllOrderTypes(t *testing.T) {
	expected := []domain.OrderType{
		domain.OrderTypeMarket,
		domain.OrderTypeLimit,
		domain.OrderTypeStopLoss,
		domain.OrderTypeStopLimit,
	}
	assert.ElementsMatch(t, expected, domain.AllOrderTypes())
}

func TestOrderType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		t    domain.OrderType
		want bool
	}{
		{"Market is valid", domain.OrderTypeMarket, true},
		{"Limit is valid", domain.OrderTypeLimit, true},
		{"StopLoss is valid", domain.OrderTypeStopLoss, true},
		{"StopLimit is valid", domain.OrderTypeStopLimit, true},
		{"Invalid value", "INVALID", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.t.IsValid())
		})
	}
}

func TestOrderType_String(t *testing.T) {
	assert.Equal(t, "MARKET", domain.OrderTypeMarket.String())
}

func TestOrderType_RequiresPrice(t *testing.T) {
	assert.True(t, domain.OrderTypeLimit.RequiresPrice())
	assert.True(t, domain.OrderTypeStopLoss.RequiresPrice())
	assert.True(t, domain.OrderTypeStopLimit.RequiresPrice())
	assert.False(t, domain.OrderTypeMarket.RequiresPrice())
	assert.False(t, domain.OrderType("INVALID").RequiresPrice())
}

func TestOrderType_IsImmediateExecution(t *testing.T) {
	assert.True(t, domain.OrderTypeMarket.IsImmediateExecution())
	assert.False(t, domain.OrderTypeLimit.IsImmediateExecution())
}

func TestOrderType_IsConditional(t *testing.T) {
	assert.True(t, domain.OrderTypeStopLoss.IsConditional())
	assert.True(t, domain.OrderTypeStopLimit.IsConditional())
	assert.False(t, domain.OrderTypeMarket.IsConditional())
}

func TestParseOrderType(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    domain.OrderType
		wantErr assert.ErrorAssertionFunc
	}{
		{"Parse MARKET", "MARKET", domain.OrderTypeMarket, assert.NoError},
		{"Parse LIMIT", "LIMIT", domain.OrderTypeLimit, assert.NoError},
		{"Parse STOP_LOSS", "STOP_LOSS", domain.OrderTypeStopLoss, assert.NoError},
		{"Parse STOP_LIMIT", "STOP_LIMIT", domain.OrderTypeStopLimit, assert.NoError},
		{"Parse invalid", "INVALID", "", func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.EqualError(t, err, "invalid order type: INVALID")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ParseOrderType(tt.s)
			tt.wantErr(t, err, fmt.Sprintf("ParseOrderType(%v)", tt.s))
			assert.Equal(t, tt.want, got, fmt.Sprintf("ParseOrderType(%v)", tt.s))
		})
	}
}

func TestOrderType_GetDescription(t *testing.T) {
	tests := []struct {
		name string
		t    domain.OrderType
		want string
	}{
		{"Market description", domain.OrderTypeMarket, "Execute immediately at current market price"},
		{"Limit description", domain.OrderTypeLimit, "Execute only at specified price or better"},
		{"StopLoss description", domain.OrderTypeStopLoss, "Trigger when price reaches specified stop price"},
		{"StopLimit description", domain.OrderTypeStopLimit, "Becomes limit order when stop price is reached"},
		{"Unknown description", "UNKNOWN", "Unknown order type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.t.GetDescription())
		})
	}
}

func TestOrderType_GetExecutionPriority(t *testing.T) {
	tests := []struct {
		name string
		t    domain.OrderType
		want int
	}{
		{"Market priority", domain.OrderTypeMarket, 1},
		{"Limit priority", domain.OrderTypeLimit, 2},
		{"StopLoss priority", domain.OrderTypeStopLoss, 3},
		{"StopLimit priority", domain.OrderTypeStopLimit, 4},
		{"Unknown priority", "UNKNOWN", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.t.GetExecutionPriority())
		})
	}
}

func TestOrderType_CanExecuteAtPrice(t *testing.T) {
	orderPrice := 100.0
	marketPrice := 100.0

	tests := []struct {
		name       string
		t          domain.OrderType
		orderPrice *float64
		marketPrice *float64
		orderSide  domain.OrderSide
		want       bool
	}{
		{"Market can always execute", domain.OrderTypeMarket, nil, nil, domain.OrderSideBuy, true},
		{"Limit buy can execute at or below market", domain.OrderTypeLimit, &orderPrice, &marketPrice, domain.OrderSideBuy, true},
		{"Limit buy cannot execute above market", domain.OrderTypeLimit, &orderPrice, func() *float64 { p := 101.0; return &p }(), domain.OrderSideBuy, false},
		{"Limit sell can execute at or above market", domain.OrderTypeLimit, &orderPrice, &marketPrice, domain.OrderSideSell, true},
		{"Limit sell cannot execute below market", domain.OrderTypeLimit, &orderPrice, func() *float64 { p := 99.0; return &p }(), domain.OrderSideSell, false},
		{"Limit with nil price", domain.OrderTypeLimit, nil, &marketPrice, domain.OrderSideBuy, false},
		{"StopLoss buy can execute at or above market", domain.OrderTypeStopLoss, &orderPrice, &marketPrice, domain.OrderSideBuy, true},
		{"StopLoss buy cannot execute below market", domain.OrderTypeStopLoss, &orderPrice, func() *float64 { p := 99.0; return &p }(), domain.OrderSideBuy, false},
		{"StopLoss sell can execute at or below market", domain.OrderTypeStopLoss, &orderPrice, &marketPrice, domain.OrderSideSell, true},
		{"StopLoss sell cannot execute above market", domain.OrderTypeStopLoss, &orderPrice, func() *float64 { p := 101.0; return &p }(), domain.OrderSideSell, false},
		{"StopLoss with nil price", domain.OrderTypeStopLoss, nil, &marketPrice, domain.OrderSideBuy, false},
		{"StopLimit not implemented", domain.OrderTypeStopLimit, nil, nil, domain.OrderSideBuy, false},
		{"Unknown type", "UNKNOWN", nil, nil, domain.OrderSideBuy, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.t.CanExecuteAtPrice(tt.orderPrice, tt.marketPrice, tt.orderSide))
		})
	}
}
