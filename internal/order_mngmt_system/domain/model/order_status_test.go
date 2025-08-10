package domain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

func TestAllOrderStatuses(t *testing.T) {
	expected := []domain.OrderStatus{
		domain.OrderStatusPending,
		domain.OrderStatusProcessing,
		domain.OrderStatusExecuted,
		domain.OrderStatusFailed,
		domain.OrderStatusCancelled,
	}
	assert.ElementsMatch(t, expected, domain.AllOrderStatuses())
}

func TestOrderStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		s      domain.OrderStatus
		want   bool
	}{
		{"Pending is valid", domain.OrderStatusPending, true},
		{"Processing is valid", domain.OrderStatusProcessing, true},
		{"Executed is valid", domain.OrderStatusExecuted, true},
		{"Failed is valid", domain.OrderStatusFailed, true},
		{"Cancelled is valid", domain.OrderStatusCancelled, true},
		{"Invalid value", "INVALID", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.IsValid())
		})
	}
}

func TestOrderStatus_String(t *testing.T) {
	assert.Equal(t, "PENDING", domain.OrderStatusPending.String())
}

func TestOrderStatus_IsTerminal(t *testing.T) {
	assert.True(t, domain.OrderStatusExecuted.IsTerminal())
	assert.True(t, domain.OrderStatusFailed.IsTerminal())
	assert.True(t, domain.OrderStatusCancelled.IsTerminal())
	assert.False(t, domain.OrderStatusPending.IsTerminal())
	assert.False(t, domain.OrderStatusProcessing.IsTerminal())
}

func TestOrderStatus_IsActive(t *testing.T) {
	assert.True(t, domain.OrderStatusPending.IsActive())
	assert.True(t, domain.OrderStatusProcessing.IsActive())
	assert.False(t, domain.OrderStatusExecuted.IsActive())
	assert.False(t, domain.OrderStatusFailed.IsActive())
	assert.False(t, domain.OrderStatusCancelled.IsActive())
}

func TestOrderStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		s      domain.OrderStatus
		target domain.OrderStatus
		want   bool
	}{
		{"Pending to Processing", domain.OrderStatusPending, domain.OrderStatusProcessing, true},
		{"Pending to Cancelled", domain.OrderStatusPending, domain.OrderStatusCancelled, true},
		{"Pending to Failed", domain.OrderStatusPending, domain.OrderStatusFailed, true},
		{"Pending to Executed", domain.OrderStatusPending, domain.OrderStatusExecuted, false},
		{"Processing to Executed", domain.OrderStatusProcessing, domain.OrderStatusExecuted, true},
		{"Processing to Failed", domain.OrderStatusProcessing, domain.OrderStatusFailed, true},
		{"Processing to Cancelled", domain.OrderStatusProcessing, domain.OrderStatusCancelled, true},
		{"Processing to Pending", domain.OrderStatusProcessing, domain.OrderStatusPending, false},
		{"Executed to any", domain.OrderStatusExecuted, domain.OrderStatusPending, false},
		{"Failed to any", domain.OrderStatusFailed, domain.OrderStatusPending, false},
		{"Cancelled to any", domain.OrderStatusCancelled, domain.OrderStatusPending, false},
		{"Invalid to any", "INVALID", domain.OrderStatusPending, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.CanTransitionTo(tt.target))
		})
	}
}

func TestParseOrderStatus(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    domain.OrderStatus
		wantErr assert.ErrorAssertionFunc
	}{
		{"Parse PENDING", "PENDING", domain.OrderStatusPending, assert.NoError},
		{"Parse PROCESSING", "PROCESSING", domain.OrderStatusProcessing, assert.NoError},
		{"Parse EXECUTED", "EXECUTED", domain.OrderStatusExecuted, assert.NoError},
		{"Parse FAILED", "FAILED", domain.OrderStatusFailed, assert.NoError},
		{"Parse CANCELLED", "CANCELLED", domain.OrderStatusCancelled, assert.NoError},
		{"Parse invalid", "INVALID", "", func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.EqualError(t, err, "invalid order status: INVALID")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ParseOrderStatus(tt.s)
			tt.wantErr(t, err, fmt.Sprintf("ParseOrderStatus(%v)", tt.s))
			assert.Equal(t, tt.want, got, fmt.Sprintf("ParseOrderStatus(%v)", tt.s))
		})
	}
}

func TestOrderStatus_GetDescription(t *testing.T) {
	tests := []struct {
		name string
		s    domain.OrderStatus
		want string
	}{
		{"Pending description", domain.OrderStatusPending, "Order submitted and waiting for processing"},
		{"Processing description", domain.OrderStatusProcessing, "Order is currently being processed"},
		{"Executed description", domain.OrderStatusExecuted, "Order has been successfully executed"},
		{"Failed description", domain.OrderStatusFailed, "Order execution failed"},
		{"Cancelled description", domain.OrderStatusCancelled, "Order has been cancelled"},
		{"Unknown description", "UNKNOWN", "Unknown status"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.GetDescription())
		})
	}
}
