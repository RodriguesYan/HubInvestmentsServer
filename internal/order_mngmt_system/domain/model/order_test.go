package domain_test

import (
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func float64Ptr(v float64) *float64 {
	return &v
}

func TestNewOrder(t *testing.T) {
	t.Run("should create a new limit buy order successfully", func(t *testing.T) {
		price := 150.0
		order, err := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "user1", order.UserID())
		assert.Equal(t, "AAPL", order.Symbol())
		assert.Equal(t, domain.OrderSideBuy, order.OrderSide())
		assert.Equal(t, domain.OrderTypeLimit, order.OrderType())
		assert.Equal(t, 10.0, order.Quantity())
		assert.Equal(t, &price, order.Price())
		assert.Equal(t, domain.OrderStatusPending, order.Status())
		assert.NotZero(t, order.ID())
		assert.NotZero(t, order.CreatedAt())
		assert.NotZero(t, order.UpdatedAt())
	})

	t.Run("should create a new market sell order successfully", func(t *testing.T) {
		order, err := domain.NewOrder("user2", "GOOGL", domain.OrderSideSell, domain.OrderTypeMarket, 5, nil)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "user2", order.UserID())
		assert.Equal(t, "GOOGL", order.Symbol())
		assert.Equal(t, domain.OrderSideSell, order.OrderSide())
		assert.Equal(t, domain.OrderTypeMarket, order.OrderType())
		assert.Equal(t, 5.0, order.Quantity())
		assert.Nil(t, order.Price())
		assert.Equal(t, domain.OrderStatusPending, order.Status())
	})

	t.Run("should return error for empty user ID", func(t *testing.T) {
		_, err := domain.NewOrder("", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		assert.Error(t, err)
		assert.Equal(t, "user ID cannot be empty", err.Error())
	})

	t.Run("should return error for empty symbol", func(t *testing.T) {
		_, err := domain.NewOrder("user1", "", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		assert.Error(t, err)
		assert.Equal(t, "symbol cannot be empty", err.Error())
	})

	t.Run("should return error for invalid order side", func(t *testing.T) {
		_, err := domain.NewOrder("user1", "AAPL", domain.OrderSide(3), domain.OrderTypeMarket, 10, nil)
		assert.Error(t, err)
		assert.Equal(t, "invalid order side", err.Error())
	})

	t.Run("should return error for invalid order type", func(t *testing.T) {
		_, err := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderType("INVALID"), 10, nil)
		assert.Error(t, err)
		assert.Equal(t, "invalid order type", err.Error())
	})

	t.Run("should return error for non-positive quantity", func(t *testing.T) {
		_, err := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 0, nil)
		assert.Error(t, err)
		assert.Equal(t, "quantity must be positive", err.Error())
	})

	t.Run("should return error for limit order without price", func(t *testing.T) {
		_, err := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10, nil)
		assert.Error(t, err)
		assert.Equal(t, "limit orders must have a price", err.Error())
	})

	t.Run("should return error for market order with price", func(t *testing.T) {
		price := 150.0
		_, err := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, &price)
		assert.Error(t, err)
		assert.Equal(t, "market orders cannot have a price", err.Error())
	})
}

func TestNewOrderFromRepository(t *testing.T) {
	now := time.Now()
	price := 150.0
	execPrice := 151.0
	marketPrice := 149.0
	marketTime := now.Add(-1 * time.Minute)

	order := domain.NewOrderFromRepository(
		"order1", "user1", "AAPL",
		domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price,
		domain.OrderStatusExecuted, now, now, &now,
		&execPrice, &marketPrice, &marketTime,
	)

	assert.NotNil(t, order)
	assert.Equal(t, "order1", order.ID())
	assert.Equal(t, "user1", order.UserID())
	assert.Equal(t, "AAPL", order.Symbol())
	assert.Equal(t, domain.OrderSideBuy, order.OrderSide())
	assert.Equal(t, domain.OrderTypeLimit, order.OrderType())
	assert.Equal(t, 10.0, order.Quantity())
	assert.Equal(t, &price, order.Price())
	assert.Equal(t, domain.OrderStatusExecuted, order.Status())
	assert.Equal(t, now, order.CreatedAt())
	assert.Equal(t, now, order.UpdatedAt())
	assert.Equal(t, &now, order.ExecutedAt())
	assert.Equal(t, &execPrice, order.ExecutionPrice())
	assert.Equal(t, &marketPrice, order.MarketPriceAtSubmission())
	assert.Equal(t, &marketTime, order.MarketDataTimestamp())
}

func TestOrder_StatusChecks(t *testing.T) {
	order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	assert.True(t, order.IsPending())
	assert.False(t, order.IsExecuted())
	assert.False(t, order.IsFailed())
	assert.False(t, order.IsCancelled())

	order.MarkAsProcessing()
	assert.False(t, order.IsPending())

	order.MarkAsExecuted(150.0)
	assert.True(t, order.IsExecuted())

	order, _ = domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
	order.MarkAsFailed()
	assert.True(t, order.IsFailed())

	order, _ = domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
	order.MarkAsCancelled()
	assert.True(t, order.IsCancelled())
}

func TestOrder_CanCancelExecute(t *testing.T) {
	order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
	assert.True(t, order.CanCancel())
	assert.True(t, order.CanExecute())

	order.MarkAsProcessing()
	assert.True(t, order.CanCancel())
	assert.True(t, order.CanExecute())

	order.MarkAsExecuted(150.0)
	assert.False(t, order.CanCancel())
	assert.False(t, order.CanExecute())

	order, _ = domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
	order.MarkAsCancelled()
	assert.False(t, order.CanCancel())
	assert.False(t, order.CanExecute())

	order, _ = domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
	order.MarkAsFailed()
	assert.False(t, order.CanCancel())
	assert.False(t, order.CanExecute()) // A failed order might be retried, so it can be executed
}

func TestOrder_SetMarketDataContext(t *testing.T) {
	order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
	marketPrice := 150.5
	timestamp := time.Now()

	order.SetMarketDataContext(marketPrice, timestamp)

	assert.Equal(t, &marketPrice, order.MarketPriceAtSubmission())
	assert.Equal(t, &timestamp, order.MarketDataTimestamp())
	assert.True(t, order.UpdatedAt().After(order.CreatedAt()))
}

func TestOrder_Marking(t *testing.T) {
	t.Run("MarkAsProcessing", func(t *testing.T) {
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		err := order.MarkAsProcessing()
		assert.NoError(t, err)
		assert.Equal(t, domain.OrderStatusProcessing, order.Status())

		order.MarkAsExecuted(150.0)
		err = order.MarkAsProcessing()
		assert.Error(t, err)
	})

	t.Run("MarkAsExecuted", func(t *testing.T) {
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		executionPrice := 152.0
		err := order.MarkAsExecuted(executionPrice)
		assert.NoError(t, err)
		assert.Equal(t, domain.OrderStatusExecuted, order.Status())
		assert.Equal(t, &executionPrice, order.ExecutionPrice())
		assert.NotNil(t, order.ExecutedAt())

		order, _ = domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		order.MarkAsCancelled()
		err = order.MarkAsExecuted(150.0)
		assert.Error(t, err)
	})

	t.Run("MarkAsFailed", func(t *testing.T) {
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		err := order.MarkAsFailed()
		assert.NoError(t, err)
		assert.Equal(t, domain.OrderStatusFailed, order.Status())

		order, _ = domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		order.MarkAsExecuted(150.0)
		err = order.MarkAsFailed()
		if assert.Error(t, err) {
			assert.Equal(t, "cannot fail an already executed order", err.Error())
		}
	})

	t.Run("MarkAsCancelled", func(t *testing.T) {
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		err := order.MarkAsCancelled()
		assert.NoError(t, err)
		assert.Equal(t, domain.OrderStatusCancelled, order.Status())

		order.MarkAsExecuted(150.0)
		err = order.MarkAsCancelled()
		assert.Error(t, err)
	})
}

func TestOrder_CalculateValues(t *testing.T) {
	price := 150.0
	limitOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)
	marketOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	assert.Equal(t, 1500.0, limitOrder.CalculateOrderValue())
	assert.Equal(t, 0.0, marketOrder.CalculateOrderValue())

	executionPrice := 152.0
	limitOrder.MarkAsExecuted(executionPrice)
	assert.Equal(t, 1520.0, limitOrder.CalculateExecutionValue())
	assert.Equal(t, 0.0, marketOrder.CalculateExecutionValue())
}

func TestOrder_GetPriceForExecution(t *testing.T) {
	limitPrice := 150.0
	marketPrice := 155.0
	limitOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &limitPrice)
	marketOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	assert.Equal(t, limitPrice, limitOrder.GetPriceForExecution(marketPrice))
	assert.Equal(t, marketPrice, marketOrder.GetPriceForExecution(marketPrice))
}

func TestOrder_ValidateForExecution(t *testing.T) {
	t.Run("valid buy limit order", func(t *testing.T) {
		price := 155.0
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)
		err := order.ValidateForExecution(150.0)
		assert.NoError(t, err)
	})

	t.Run("buy limit price too high", func(t *testing.T) {
		price := 170.0
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)
		err := order.ValidateForExecution(150.0)
		assert.Error(t, err)
		assert.Equal(t, "buy limit price too far above market price", err.Error())
	})

	t.Run("valid sell limit order", func(t *testing.T) {
		price := 145.0
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideSell, domain.OrderTypeLimit, 10, &price)
		err := order.ValidateForExecution(150.0)
		assert.NoError(t, err)
	})

	t.Run("sell limit price too low", func(t *testing.T) {
		price := 130.0
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideSell, domain.OrderTypeLimit, 10, &price)
		err := order.ValidateForExecution(150.0)
		assert.Error(t, err)
		assert.Equal(t, "sell limit price too far below market price", err.Error())
	})

	t.Run("market order should pass", func(t *testing.T) {
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		err := order.ValidateForExecution(150.0)
		assert.NoError(t, err)
	})

	t.Run("cannot execute order in wrong status", func(t *testing.T) {
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
		order.MarkAsExecuted(150.0)
		err := order.ValidateForExecution(150.0)
		assert.Error(t, err)
	})
}

func TestOrder_ValidatePositionForSellOrder(t *testing.T) {
	sellOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideSell, domain.OrderTypeMarket, 10, nil)
	buyOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)

	t.Run("sufficient position", func(t *testing.T) {
		err := sellOrder.ValidatePositionForSellOrder(15.0)
		assert.NoError(t, err)
	})

	t.Run("insufficient position", func(t *testing.T) {
		err := sellOrder.ValidatePositionForSellOrder(5.0)
		assert.Error(t, err)
		assert.Equal(t, "insufficient position: cannot sell more than available quantity", err.Error())
	})

	t.Run("no position", func(t *testing.T) {
		err := sellOrder.ValidatePositionForSellOrder(0)
		assert.Error(t, err)
		assert.Equal(t, "no position available for this symbol", err.Error())
	})

	t.Run("buy order should not be validated", func(t *testing.T) {
		err := buyOrder.ValidatePositionForSellOrder(5.0)
		assert.NoError(t, err)
	})
}

func TestOrder_Validate(t *testing.T) {
	t.Run("valid order", func(t *testing.T) {
		price := 150.0
		order, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)
		err := order.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid user id", func(t *testing.T) {
		order := domain.NewOrderFromRepository("id", "", "sym", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
		err := order.Validate()
		assert.ErrorContains(t, err, "user ID cannot be empty")
	})

	t.Run("invalid symbol", func(t *testing.T) {
		order := domain.NewOrderFromRepository("id", "user", "", domain.OrderSideBuy, domain.OrderTypeMarket, 1, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
		err := order.Validate()
		assert.ErrorContains(t, err, "symbol cannot be empty")
	})

	t.Run("invalid order side", func(t *testing.T) {
		order := domain.NewOrderFromRepository("id", "user", "sym", 3, domain.OrderTypeMarket, 1, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
		err := order.Validate()
		assert.ErrorContains(t, err, "invalid order side")
	})

	t.Run("invalid order type", func(t *testing.T) {
		order := domain.NewOrderFromRepository("id", "user", "sym", domain.OrderSideBuy, "invalid", 1, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
		err := order.Validate()
		assert.ErrorContains(t, err, "invalid order type")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		order := domain.NewOrderFromRepository("id", "user", "sym", domain.OrderSideBuy, domain.OrderTypeMarket, 0, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
		err := order.Validate()
		assert.ErrorContains(t, err, "quantity must be positive")
	})

	t.Run("limit order with nil price", func(t *testing.T) {
		order := domain.NewOrderFromRepository("id", "user", "sym", domain.OrderSideBuy, domain.OrderTypeLimit, 1, nil, domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
		err := order.Validate()
		assert.ErrorContains(t, err, "limit orders must have a price")
	})

	t.Run("limit order with zero price", func(t *testing.T) {
		order := domain.NewOrderFromRepository("id", "user", "sym", domain.OrderSideBuy, domain.OrderTypeLimit, 1, float64Ptr(0), domain.OrderStatusPending, time.Now(), time.Now(), nil, nil, nil, nil)
		err := order.Validate()
		assert.ErrorContains(t, err, "limit price must be positive")
	})
}

func TestOrder_GetOrderDescription(t *testing.T) {
	price := 150.0
	limitOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeLimit, 10, &price)
	marketOrder, _ := domain.NewOrder("user1", "GOOGL", domain.OrderSideSell, domain.OrderTypeMarket, 5, nil)

	assert.Equal(t, "Buy order - purchasing assets Execute only at specified price or better of 10.00 AAPL at 150.00", limitOrder.GetOrderDescription())
	assert.Equal(t, "Sell order - selling assets from portfolio Execute immediately at current market price of 5.00 GOOGL", marketOrder.GetOrderDescription())
}

func TestOrder_SideMethods(t *testing.T) {
	buyOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideBuy, domain.OrderTypeMarket, 10, nil)
	sellOrder, _ := domain.NewOrder("user1", "AAPL", domain.OrderSideSell, domain.OrderTypeMarket, 10, nil)

	assert.True(t, buyOrder.IsBuyOrder())
	assert.False(t, buyOrder.IsSellOrder())
	assert.False(t, buyOrder.RequiresPositionValidation())

	assert.False(t, sellOrder.IsBuyOrder())
	assert.True(t, sellOrder.IsSellOrder())
	assert.True(t, sellOrder.RequiresPositionValidation())
}
