package repository

import (
	"context"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// IOrderRepository defines the contract for order persistence operations
type IOrderRepository interface {
	// Save saves a new order to the database
	Save(ctx context.Context, order *domain.Order) error

	// FindByID retrieves an order by its unique identifier
	FindByID(ctx context.Context, orderID string) (*domain.Order, error)

	// FindByUserID retrieves all orders for a specific user
	FindByUserID(ctx context.Context, userID string) ([]*domain.Order, error)

	// UpdateStatus updates the status of an existing order
	UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error

	// UpdateExecutionDetails updates order with execution details
	UpdateExecutionDetails(ctx context.Context, orderID string, executionPrice float64, executedAt time.Time) error

	// FindByUserIDAndStatus retrieves orders for a user filtered by status
	FindByUserIDAndStatus(ctx context.Context, userID string, status domain.OrderStatus) ([]*domain.Order, error)

	// FindByStatus retrieves all orders with a specific status
	FindByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error)

	// FindOrderHistory retrieves paginated order history for a user
	FindOrderHistory(ctx context.Context, userID string, limit int, offset int) ([]*domain.Order, error)

	// FindOrdersBySymbol retrieves all orders for a specific symbol
	FindOrdersBySymbol(ctx context.Context, symbol string) ([]*domain.Order, error)

	// FindOrdersByDateRange retrieves orders within a date range
	FindOrdersByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*domain.Order, error)

	// CountOrdersByUserID returns the total number of orders for a user
	CountOrdersByUserID(ctx context.Context, userID string) (int, error)

	// Delete removes an order from the database
	Delete(ctx context.Context, orderID string) error
}
