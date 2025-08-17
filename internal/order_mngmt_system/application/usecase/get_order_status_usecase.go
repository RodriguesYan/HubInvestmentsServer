package usecase

import (
	"context"
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/internal/order_mngmt_system/domain/repository"
	"HubInvestments/internal/order_mngmt_system/infra/external"
)

type IGetOrderStatusUseCase interface {
	Execute(ctx context.Context, orderID, userID string) (*OrderStatusResult, error)
	GetOrderHistory(ctx context.Context, userID string, options *OrderHistoryOptions) (*OrderHistoryResult, error)
}

type GetOrderStatusUseCase struct {
	orderRepository  repository.IOrderRepository
	marketDataClient external.IMarketDataClient
}

type OrderStatusResult struct {
	OrderID                 string     `json:"order_id"`
	UserID                  string     `json:"user_id"`
	Symbol                  string     `json:"symbol"`
	OrderSide               string     `json:"order_side"`
	OrderType               string     `json:"order_type"`
	Quantity                float64    `json:"quantity"`
	Price                   *float64   `json:"price,omitempty"`
	Status                  string     `json:"status"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	ExecutedAt              *time.Time `json:"executed_at,omitempty"`
	ExecutionPrice          *float64   `json:"execution_price,omitempty"`
	MarketPriceAtSubmission *float64   `json:"market_price_at_submission,omitempty"`
	CurrentMarketPrice      *float64   `json:"current_market_price,omitempty"`
	PriceChange             *float64   `json:"price_change,omitempty"`
	PriceChangePercent      *float64   `json:"price_change_percent,omitempty"`
	EstimatedValue          *float64   `json:"estimated_value,omitempty"`
	StatusDescription       string     `json:"status_description"`
	CanCancel               bool       `json:"can_cancel"`
	MarketDataTimestamp     *time.Time `json:"market_data_timestamp,omitempty"`
}

type OrderHistoryOptions struct {
	Limit     int                  `json:"limit,omitempty"`
	Offset    int                  `json:"offset,omitempty"`
	Status    []domain.OrderStatus `json:"status,omitempty"`
	Symbol    string               `json:"symbol,omitempty"`
	OrderSide *domain.OrderSide    `json:"order_side,omitempty"`
	OrderType *domain.OrderType    `json:"order_type,omitempty"`
	StartDate *time.Time           `json:"start_date,omitempty"`
	EndDate   *time.Time           `json:"end_date,omitempty"`
	SortBy    string               `json:"sort_by,omitempty"`    // "created_at", "updated_at", "symbol"
	SortOrder string               `json:"sort_order,omitempty"` // "asc", "desc"
}

type OrderHistoryResult struct {
	Orders     []*OrderStatusResult `json:"orders"`
	TotalCount int                  `json:"total_count"`
	HasMore    bool                 `json:"has_more"`
	Pagination *PaginationInfo      `json:"pagination"`
}

type PaginationInfo struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	PageSize    int `json:"page_size"`
	TotalItems  int `json:"total_items"`
}

func NewGetOrderStatusUseCase(
	orderRepository repository.IOrderRepository,
	marketDataClient external.IMarketDataClient,
) IGetOrderStatusUseCase {
	return &GetOrderStatusUseCase{
		orderRepository:  orderRepository,
		marketDataClient: marketDataClient,
	}
}

// Execute retrieves the status of a specific order with market data context
func (uc *GetOrderStatusUseCase) Execute(ctx context.Context, orderID, userID string) (*OrderStatusResult, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	order, err := uc.orderRepository.FindByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	// Step 3: Verify order belongs to the user
	if order.UserID() != userID {
		return nil, fmt.Errorf("order not found")
	}

	currentMarketData, err := uc.getCurrentMarketData(ctx, order.Symbol())
	if err != nil {
		//put log in here
		currentMarketData = nil
	}

	result := uc.buildOrderStatusResult(order, currentMarketData)
	return result, nil
}

func (uc *GetOrderStatusUseCase) GetOrderHistory(ctx context.Context, userID string, options *OrderHistoryOptions) (*OrderHistoryResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if options == nil {
		options = &OrderHistoryOptions{
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
	}

	if options.Limit <= 0 || options.Limit > 100 {
		options.Limit = 50
	}
	if options.Offset < 0 {
		options.Offset = 0
	}
	if options.SortBy == "" {
		options.SortBy = "created_at"
	}
	if options.SortOrder == "" {
		options.SortOrder = "desc"
	}

	orders, err := uc.orderRepository.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order history: %w", err)
	}

	filteredOrders := uc.applyBasicFiltering(orders, options)
	totalCount := len(filteredOrders)

	start := options.Offset
	end := start + options.Limit
	if start > len(filteredOrders) {
		start = len(filteredOrders)
	}
	if end > len(filteredOrders) {
		end = len(filteredOrders)
	}

	paginatedOrders := filteredOrders[start:end]

	// Build results with market data context
	results := make([]*OrderStatusResult, 0, len(paginatedOrders))
	for _, order := range paginatedOrders {
		// Get current market data for each symbol (could be optimized with batch requests)
		currentMarketData, _ := uc.getCurrentMarketData(ctx, order.Symbol())
		result := uc.buildOrderStatusResult(order, currentMarketData)
		results = append(results, result)
	}

	pagination := uc.buildPaginationInfo(options, totalCount)

	return &OrderHistoryResult{
		Orders:     results,
		TotalCount: totalCount,
		HasMore:    options.Offset+len(paginatedOrders) < totalCount,
		Pagination: pagination,
	}, nil
}

type OrderStatusMarketDataContext struct {
	CurrentPrice float64
	Timestamp    time.Time
}

func (uc *GetOrderStatusUseCase) getCurrentMarketData(ctx context.Context, symbol string) (*OrderStatusMarketDataContext, error) {
	currentPrice, err := uc.marketDataClient.GetCurrentPrice(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get current price: %w", err)
	}

	return &OrderStatusMarketDataContext{
		CurrentPrice: currentPrice,
		Timestamp:    time.Now(),
	}, nil
}

func (uc *GetOrderStatusUseCase) buildOrderStatusResult(order *domain.Order, marketData *OrderStatusMarketDataContext) *OrderStatusResult {
	result := &OrderStatusResult{
		OrderID:                 order.ID(),
		UserID:                  order.UserID(),
		Symbol:                  order.Symbol(),
		OrderSide:               string(order.OrderSide()),
		OrderType:               string(order.OrderType()),
		Quantity:                order.Quantity(),
		Price:                   order.Price(),
		Status:                  string(order.Status()),
		CreatedAt:               order.CreatedAt(),
		UpdatedAt:               order.UpdatedAt(),
		ExecutedAt:              order.ExecutedAt(),
		ExecutionPrice:          order.ExecutionPrice(),
		MarketPriceAtSubmission: order.MarketPriceAtSubmission(),
		StatusDescription:       uc.getStatusDescription(order),
		CanCancel:               order.CanCancel(),
		MarketDataTimestamp:     order.MarketDataTimestamp(),
	}

	if marketData == nil {
		return result
	}

	result.CurrentMarketPrice = &marketData.CurrentPrice

	// Calculate price change if we have submission price
	if order.MarketPriceAtSubmission() != nil {
		submissionPrice := *order.MarketPriceAtSubmission()
		priceChange := marketData.CurrentPrice - submissionPrice
		priceChangePercent := (priceChange / submissionPrice) * 100

		result.PriceChange = &priceChange
		result.PriceChangePercent = &priceChangePercent
	}

	// Calculate estimated current value
	estimatedValue := marketData.CurrentPrice * order.Quantity()
	result.EstimatedValue = &estimatedValue

	return result
}

func (uc *GetOrderStatusUseCase) getStatusDescription(order *domain.Order) string {
	switch order.Status() {
	case domain.OrderStatusPending:
		return "Order is pending and waiting to be processed"
	case domain.OrderStatusProcessing:
		return "Order is currently being processed"
	case domain.OrderStatusExecuted:
		if order.ExecutedAt() != nil {
			return fmt.Sprintf("Order was executed on %s", order.ExecutedAt().Format("2006-01-02 15:04:05"))
		}
		return "Order has been executed successfully"
	case domain.OrderStatusFailed:
		return "Order execution failed"
	case domain.OrderStatusCancelled:
		return "Order has been cancelled"
	default:
		return "Unknown order status"
	}
}

func (uc *GetOrderStatusUseCase) applyBasicFiltering(orders []*domain.Order, options *OrderHistoryOptions) []*domain.Order {
	filtered := make([]*domain.Order, 0, len(orders))

	for _, order := range orders {
		//TODO: improve this code block
		if len(options.Status) > 0 {
			statusMatch := false
			for _, status := range options.Status {
				if order.Status() == status {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				continue
			}
		}

		// Apply symbol filter
		if options.Symbol != "" && order.Symbol() != options.Symbol {
			continue
		}

		if options.OrderSide != nil && order.OrderSide() != *options.OrderSide {
			continue
		}

		if options.OrderType != nil && order.OrderType() != *options.OrderType {
			continue
		}

		if options.StartDate != nil && order.CreatedAt().Before(*options.StartDate) {
			continue
		}
		if options.EndDate != nil && order.CreatedAt().After(*options.EndDate) {
			continue
		}

		filtered = append(filtered, order)
	}

	return filtered
}

func (uc *GetOrderStatusUseCase) buildPaginationInfo(options *OrderHistoryOptions, totalCount int) *PaginationInfo {
	currentPage := (options.Offset / options.Limit) + 1
	totalPages := (totalCount + options.Limit - 1) / options.Limit // Ceiling division

	return &PaginationInfo{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		PageSize:    options.Limit,
		TotalItems:  totalCount,
	}
}
