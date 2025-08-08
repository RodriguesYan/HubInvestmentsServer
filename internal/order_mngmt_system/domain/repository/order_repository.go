package repository

import (
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

// IOrderRepository defines the contract for order persistence operations
type IOrderRepository interface {
	// Save saves a new order to the database
	Save(order *domain.Order) error

	// FindByID retrieves an order by its unique identifier
	FindByID(orderID string) (*domain.Order, error)

	// FindByUserID retrieves all orders for a specific user
	FindByUserID(userID string) ([]*domain.Order, error)

	// UpdateStatus updates the status of an existing order
	UpdateStatus(orderID string, status domain.OrderStatus) error

	// UpdateOrderWithExecution updates order with execution details
	UpdateOrderWithExecution(orderID string, executionPrice float64, executedAt time.Time) error

	// FindByUserIDAndStatus retrieves orders for a user filtered by status
	FindByUserIDAndStatus(userID string, status domain.OrderStatus) ([]*domain.Order, error)

	// FindByUserIDAndSymbol retrieves orders for a user filtered by symbol
	FindByUserIDAndSymbol(userID string, symbol string) ([]*domain.Order, error)

	// FindByUserIDWithPagination retrieves orders for a user with pagination support
	FindByUserIDWithPagination(userID string, limit, offset int) ([]*domain.Order, error)

	// FindByUserIDAndDateRange retrieves orders for a user within a date range
	FindByUserIDAndDateRange(userID string, startDate, endDate time.Time) ([]*domain.Order, error)

	// FindActiveOrdersByUserID retrieves only active orders (PENDING, PROCESSING) for a user
	FindActiveOrdersByUserID(userID string) ([]*domain.Order, error)

	// FindActiveOrdersByUserIDAndSymbol retrieves active orders for a specific symbol
	FindActiveOrdersByUserIDAndSymbol(userID string, symbol string) ([]*domain.Order, error)

	// CountOrdersByUserID returns the total count of orders for a user
	CountOrdersByUserID(userID string) (int, error)

	// CountOrdersByUserIDAndStatus returns the count of orders for a user filtered by status
	CountOrdersByUserIDAndStatus(userID string, status domain.OrderStatus) (int, error)

	// FindOrdersForProcessing retrieves orders that are ready for processing
	FindOrdersForProcessing(limit int) ([]*domain.Order, error)

	// FindExpiredOrders retrieves orders that have exceeded their time limits
	FindExpiredOrders(expiredBefore time.Time) ([]*domain.Order, error)

	// Delete removes an order from the database (for testing/admin purposes)
	Delete(orderID string) error

	// ExistsOrderByID checks if an order exists by its ID
	ExistsOrderByID(orderID string) (bool, error)

	// FindOrdersNeedingCancellation retrieves orders that should be auto-cancelled
	FindOrdersNeedingCancellation(beforeTime time.Time) ([]*domain.Order, error)
}

// OrderQueryOptions provides flexible query options for advanced order filtering
type OrderQueryOptions struct {
	UserID      string
	Symbol      *string
	Status      *domain.OrderStatus
	OrderSide   *domain.OrderSide
	OrderType   *domain.OrderType
	StartDate   *time.Time
	EndDate     *time.Time
	MinQuantity *float64
	MaxQuantity *float64
	MinPrice    *float64
	MaxPrice    *float64
	Limit       *int
	Offset      *int
	SortBy      *string // "created_at", "updated_at", "symbol", "quantity", "price"
	SortOrder   *string // "ASC", "DESC"
}

// IOrderRepositoryAdvanced extends IOrderRepository with advanced query capabilities
type IOrderRepositoryAdvanced interface {
	IOrderRepository

	// FindOrdersWithOptions provides flexible querying with multiple filters
	FindOrdersWithOptions(options OrderQueryOptions) ([]*domain.Order, error)

	// CountOrdersWithOptions returns count of orders matching the query options
	CountOrdersWithOptions(options OrderQueryOptions) (int, error)

	// GetOrderStatistics returns aggregated statistics for orders
	GetOrderStatistics(userID string, startDate, endDate time.Time) (*OrderStatistics, error)

	// FindOrdersByIDs retrieves multiple orders by their IDs
	FindOrdersByIDs(orderIDs []string) ([]*domain.Order, error)

	// UpdateOrderMarketData updates market data context for an order
	UpdateOrderMarketData(orderID string, marketPrice float64, timestamp time.Time) error
}

// OrderStatistics provides aggregated order statistics
type OrderStatistics struct {
	TotalOrders       int     `json:"totalOrders"`
	ExecutedOrders    int     `json:"executedOrders"`
	CancelledOrders   int     `json:"cancelledOrders"`
	FailedOrders      int     `json:"failedOrders"`
	PendingOrders     int     `json:"pendingOrders"`
	ProcessingOrders  int     `json:"processingOrders"`
	TotalVolume       float64 `json:"totalVolume"`       // Total quantity traded
	TotalValue        float64 `json:"totalValue"`        // Total value of executed orders
	AverageOrderSize  float64 `json:"averageOrderSize"`  // Average order quantity
	AverageOrderValue float64 `json:"averageOrderValue"` // Average order value
	BuyOrders         int     `json:"buyOrders"`
	SellOrders        int     `json:"sellOrders"`
	UniqueSymbols     int     `json:"uniqueSymbols"`
	ExecutionRate     float64 `json:"executionRate"` // Percentage of orders executed
}

// OrderFilter provides a builder pattern for creating complex queries
type OrderFilter struct {
	conditions []string
	args       []interface{}
	limit      *int
	offset     *int
	orderBy    *string
}

// NewOrderFilter creates a new order filter builder
func NewOrderFilter() *OrderFilter {
	return &OrderFilter{
		conditions: make([]string, 0),
		args:       make([]interface{}, 0),
	}
}

// WithUserID adds user ID filter
func (f *OrderFilter) WithUserID(userID string) *OrderFilter {
	f.conditions = append(f.conditions, "user_id = ?")
	f.args = append(f.args, userID)
	return f
}

// WithStatus adds status filter
func (f *OrderFilter) WithStatus(status domain.OrderStatus) *OrderFilter {
	f.conditions = append(f.conditions, "status = ?")
	f.args = append(f.args, status)
	return f
}

// WithSymbol adds symbol filter
func (f *OrderFilter) WithSymbol(symbol string) *OrderFilter {
	f.conditions = append(f.conditions, "symbol = ?")
	f.args = append(f.args, symbol)
	return f
}

// WithOrderSide adds order side filter
func (f *OrderFilter) WithOrderSide(side domain.OrderSide) *OrderFilter {
	f.conditions = append(f.conditions, "order_side = ?")
	f.args = append(f.args, side)
	return f
}

// WithDateRange adds date range filter
func (f *OrderFilter) WithDateRange(startDate, endDate time.Time) *OrderFilter {
	f.conditions = append(f.conditions, "created_at BETWEEN ? AND ?")
	f.args = append(f.args, startDate, endDate)
	return f
}

// WithPagination adds limit and offset
func (f *OrderFilter) WithPagination(limit, offset int) *OrderFilter {
	f.limit = &limit
	f.offset = &offset
	return f
}

// WithOrderBy adds ordering
func (f *OrderFilter) WithOrderBy(column, direction string) *OrderFilter {
	orderBy := column + " " + direction
	f.orderBy = &orderBy
	return f
}

// Build returns the filter conditions and arguments
func (f *OrderFilter) Build() ([]string, []interface{}, *int, *int, *string) {
	return f.conditions, f.args, f.limit, f.offset, f.orderBy
}
