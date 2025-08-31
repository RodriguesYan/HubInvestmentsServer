package dto

import (
	"fmt"
	"strconv"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"

	"github.com/google/uuid"
)

type OrderMapper struct{}

func NewOrderMapper() *OrderMapper {
	return &OrderMapper{}
}

// ToDTO converts domain Order to OrderDTO
func (m *OrderMapper) ToDTO(order *domain.Order) (*OrderDTO, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	orderUUID, err := uuid.Parse(order.ID())
	if err != nil {
		return nil, fmt.Errorf("invalid order ID format: %w", err)
	}

	userID, err := strconv.Atoi(order.UserID())
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	dto := &OrderDTO{
		ID:        orderUUID,
		UserID:    userID,
		Symbol:    order.Symbol(),
		OrderType: order.OrderType().String(),
		OrderSide: order.OrderSide().String(),
		Quantity:  order.Quantity(),
		Price:     order.Price(),
		Status:    order.Status().String(),
		CreatedAt: order.CreatedAt(),
		UpdatedAt: order.UpdatedAt(),
	}

	// Handle optional execution fields
	if order.ExecutedAt() != nil {
		dto.ExecutedAt = order.ExecutedAt()
	}

	if order.ExecutionPrice() != nil {
		dto.ExecutionPrice = order.ExecutionPrice()
	}

	if order.MarketPriceAtSubmission() != nil {
		dto.MarketPriceAtSubmission = order.MarketPriceAtSubmission()
	}

	if order.MarketDataTimestamp() != nil {
		dto.MarketDataTimestamp = order.MarketDataTimestamp()
	}

	return dto, nil
}

func (m *OrderMapper) ToDomain(dto *OrderDTO) (*domain.Order, error) {
	if dto == nil {
		return nil, fmt.Errorf("dto cannot be nil")
	}

	// Parse order type
	orderType, err := m.parseOrderType(dto.OrderType)
	if err != nil {
		return nil, fmt.Errorf("invalid order type: %w", err)
	}

	orderSide, err := m.parseOrderSide(dto.OrderSide)
	if err != nil {
		return nil, fmt.Errorf("invalid order side: %w", err)
	}

	orderStatus, err := m.parseOrderStatus(dto.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid order status: %w", err)
	}

	orderID := dto.ID.String()
	userID := strconv.Itoa(dto.UserID)

	order := domain.NewOrderFromDatabase(
		orderID,
		userID,
		dto.Symbol,
		orderSide,
		orderType,
		dto.Quantity,
		dto.Price,
		orderStatus,
		dto.CreatedAt,
		dto.UpdatedAt,
		dto.ExecutedAt,
		dto.ExecutionPrice,
		dto.MarketPriceAtSubmission,
		dto.MarketDataTimestamp,
	)

	return order, nil
}

func (m *OrderMapper) parseOrderType(typeStr string) (domain.OrderType, error) {
	switch typeStr {
	case "MARKET":
		return domain.OrderTypeMarket, nil
	case "LIMIT":
		return domain.OrderTypeLimit, nil
	case "STOP_LOSS":
		return domain.OrderTypeStopLoss, nil
	case "STOP_LIMIT":
		return domain.OrderTypeStopLimit, nil
	default:
		return "", fmt.Errorf("unknown order type: %s", typeStr)
	}
}

func (m *OrderMapper) parseOrderSide(sideStr string) (domain.OrderSide, error) {
	switch sideStr {
	case "BUY":
		return domain.OrderSideBuy, nil
	case "SELL":
		return domain.OrderSideSell, nil
	default:
		return domain.OrderSide(0), fmt.Errorf("unknown order side: %s", sideStr)
	}
}

func (m *OrderMapper) parseOrderStatus(statusStr string) (domain.OrderStatus, error) {
	switch statusStr {
	case "PENDING":
		return domain.OrderStatusPending, nil
	case "PROCESSING":
		return domain.OrderStatusProcessing, nil
	case "EXECUTED":
		return domain.OrderStatusExecuted, nil
	case "FAILED":
		return domain.OrderStatusFailed, nil
	case "CANCELLED":
		return domain.OrderStatusCancelled, nil
	default:
		return "", fmt.Errorf("unknown order status: %s", statusStr)
	}
}

func (m *OrderMapper) ToOrderList(dtos []*OrderDTO) ([]*domain.Order, error) {
	if dtos == nil {
		return nil, nil
	}

	orders := make([]*domain.Order, len(dtos))
	for i, dto := range dtos {
		order, err := m.ToDomain(dto)
		if err != nil {
			return nil, fmt.Errorf("failed to convert order at index %d: %w", i, err)
		}
		orders[i] = order
	}

	return orders, nil
}

func ParseUserIDFromString(userIDStr string) (int, error) {
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}
	return userID, nil
}

func ParseUUIDFromString(uuidStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format: %w", err)
	}
	return id, nil
}
