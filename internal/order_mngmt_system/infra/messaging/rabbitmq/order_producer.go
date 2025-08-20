package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/shared/infra/messaging"
)

type OrderMessage struct {
	OrderID                 string               `json:"order_id"`
	UserID                  string               `json:"user_id"`
	Symbol                  string               `json:"symbol"`
	OrderSide               string               `json:"order_side"`
	OrderType               string               `json:"order_type"`
	Quantity                float64              `json:"quantity"`
	Price                   *float64             `json:"price,omitempty"`
	Status                  string               `json:"status"`
	CreatedAt               time.Time            `json:"created_at"`
	UpdatedAt               time.Time            `json:"updated_at"`
	ExecutedAt              *time.Time           `json:"executed_at,omitempty"`
	ExecutionPrice          *float64             `json:"execution_price,omitempty"`
	MarketPriceAtSubmission *float64             `json:"market_price_at_submission,omitempty"`
	MarketDataTimestamp     *time.Time           `json:"market_data_timestamp,omitempty"`
	MessageMetadata         OrderMessageMetadata `json:"message_metadata"`
}

type OrderMessageMetadata struct {
	MessageID       string    `json:"message_id"`
	CorrelationID   string    `json:"correlation_id"`
	Timestamp       time.Time `json:"timestamp"`
	RetryAttempt    int       `json:"retry_attempt"`
	Priority        uint8     `json:"priority"`
	Source          string    `json:"source"`
	MessageType     string    `json:"message_type"`
	ProcessingStage string    `json:"processing_stage"`
}

type OrderProducer struct {
	queueManager   *OrderQueueManager
	messageHandler messaging.MessageHandler
}

func NewOrderProducer(messageHandler messaging.MessageHandler) *OrderProducer {
	return &OrderProducer{
		queueManager:   NewOrderQueueManager(messageHandler),
		messageHandler: messageHandler,
	}
}

func NewOrderProducerWithQueueManager(queueManager *OrderQueueManager) *OrderProducer {
	return &OrderProducer{
		queueManager:   queueManager,
		messageHandler: queueManager.messageHandler,
	}
}

// PublishOrderForProcessing is the main method for sending orders for asynchronous processing
func (op *OrderProducer) PublishOrderForProcessing(ctx context.Context, order *domain.Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}

	orderMessage, err := op.createOrderMessage(order, "order_processing", "processing")
	if err != nil {
		return fmt.Errorf("failed to create order message: %w", err)
	}

	messageBytes, err := json.Marshal(orderMessage)
	if err != nil {
		return fmt.Errorf("failed to serialize order message: %w", err)
	}

	priority := op.calculateMessagePriority(order)

	err = op.queueManager.PublishToProcessingQueue(ctx, messageBytes, orderMessage.MessageMetadata.MessageID, priority)
	if err != nil {
		return fmt.Errorf("failed to publish order to processing queue: %w", err)
	}

	return nil
}

func (op *OrderProducer) PublishOrderForSubmission(ctx context.Context, order *domain.Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}

	orderMessage, err := op.createOrderMessage(order, "order_submission", "submission")
	if err != nil {
		return fmt.Errorf("failed to create order message: %w", err)
	}

	messageBytes, err := json.Marshal(orderMessage)
	if err != nil {
		return fmt.Errorf("failed to serialize order message: %w", err)
	}

	err = op.queueManager.PublishToSubmitQueue(ctx, messageBytes, orderMessage.MessageMetadata.MessageID)
	if err != nil {
		return fmt.Errorf("failed to publish order to submission queue: %w", err)
	}

	return nil
}

func (op *OrderProducer) PublishOrderForRetry(ctx context.Context, order *domain.Order, retryAttempt int, reason string) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}

	if retryAttempt < 0 {
		return fmt.Errorf("retry attempt must be non-negative")
	}

	orderMessage, err := op.createOrderMessage(order, "order_retry", "retry")
	if err != nil {
		return fmt.Errorf("failed to create order message: %w", err)
	}

	orderMessage.MessageMetadata.RetryAttempt = retryAttempt
	orderMessage.MessageMetadata.MessageType = "order_retry"

	// Include retry reason in correlation ID for tracking failed processing attempts
	if reason != "" {
		orderMessage.MessageMetadata.CorrelationID = fmt.Sprintf("%s_retry_%d_%s",
			order.ID(), retryAttempt, reason)
	}

	messageBytes, err := json.Marshal(orderMessage)
	if err != nil {
		return fmt.Errorf("failed to serialize retry order message: %w", err)
	}

	err = op.queueManager.PublishToRetryQueue(ctx, messageBytes, orderMessage.MessageMetadata.MessageID, retryAttempt)
	if err != nil {
		return fmt.Errorf("failed to publish order to retry queue: %w", err)
	}

	return nil
}

func (op *OrderProducer) PublishOrderStatusUpdate(ctx context.Context, order *domain.Order, previousStatus string) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}

	statusUpdate := OrderStatusUpdate{
		OrderID:        order.ID(),
		UserID:         order.UserID(),
		PreviousStatus: previousStatus,
		CurrentStatus:  string(order.Status()),
		UpdatedAt:      order.UpdatedAt(),
		ExecutedAt:     order.ExecutedAt(),
		ExecutionPrice: order.ExecutionPrice(),
		MessageMetadata: OrderMessageMetadata{
			MessageID:       fmt.Sprintf("status_%s_%d", order.ID(), time.Now().UnixNano()),
			CorrelationID:   order.ID(),
			Timestamp:       time.Now(),
			RetryAttempt:    0,
			Priority:        8, // High priority ensures status updates are processed before regular orders
			Source:          "order_producer",
			MessageType:     "status_update",
			ProcessingStage: "status_notification",
		},
	}

	messageBytes, err := json.Marshal(statusUpdate)
	if err != nil {
		return fmt.Errorf("failed to serialize status update message: %w", err)
	}

	err = op.queueManager.PublishStatusUpdate(ctx, messageBytes, order.ID())
	if err != nil {
		return fmt.Errorf("failed to publish status update: %w", err)
	}

	return nil
}

func (op *OrderProducer) PublishBatchOrders(ctx context.Context, orders []*domain.Order) error {
	if len(orders) == 0 {
		return fmt.Errorf("orders list cannot be empty")
	}

	// Validate all orders before publishing any to ensure atomicity
	for i, order := range orders {
		if order == nil {
			return fmt.Errorf("order at index %d cannot be nil", i)
		}
	}

	batchID := fmt.Sprintf("batch_%d_%d", time.Now().Unix(), len(orders))

	var publishedCount int
	for i, order := range orders {
		orderMessage, err := op.createOrderMessage(order, "batch_order_processing", "batch_processing")
		if err != nil {
			return fmt.Errorf("failed to create message for order %d: %w", i, err)
		}

		orderMessage.MessageMetadata.CorrelationID = fmt.Sprintf("%s_order_%d", batchID, i)

		messageBytes, err := json.Marshal(orderMessage)
		if err != nil {
			return fmt.Errorf("failed to serialize order %d: %w", i, err)
		}

		priority := op.calculateMessagePriority(order)
		err = op.queueManager.PublishToProcessingQueue(ctx, messageBytes, orderMessage.MessageMetadata.MessageID, priority)
		if err != nil {
			return fmt.Errorf("failed to publish order %d (published %d/%d): %w", i, publishedCount, len(orders), err)
		}

		publishedCount++
	}

	return nil
}

func (op *OrderProducer) createOrderMessage(order *domain.Order, messageType, processingStage string) (*OrderMessage, error) {
	messageID := fmt.Sprintf("%s_%s_%d", messageType, order.ID(), time.Now().UnixNano())

	orderMessage := &OrderMessage{
		OrderID:                 order.ID(),
		UserID:                  order.UserID(),
		Symbol:                  order.Symbol(),
		OrderSide:               order.OrderSide().String(),
		OrderType:               order.OrderType().String(),
		Quantity:                order.Quantity(),
		Price:                   order.Price(),
		Status:                  order.Status().String(),
		CreatedAt:               order.CreatedAt(),
		UpdatedAt:               order.UpdatedAt(),
		ExecutedAt:              order.ExecutedAt(),
		ExecutionPrice:          order.ExecutionPrice(),
		MarketPriceAtSubmission: order.MarketPriceAtSubmission(),
		MarketDataTimestamp:     order.MarketDataTimestamp(),
		MessageMetadata: OrderMessageMetadata{
			MessageID:       messageID,
			CorrelationID:   order.ID(),
			Timestamp:       time.Now(),
			RetryAttempt:    0,
			Priority:        op.calculateMessagePriority(order),
			Source:          "order_producer",
			MessageType:     messageType,
			ProcessingStage: processingStage,
		},
	}

	return orderMessage, nil
}

// calculateMessagePriority determines message priority based on order characteristics
// Higher priority orders are processed first to minimize market impact and risk
func (op *OrderProducer) calculateMessagePriority(order *domain.Order) uint8 {
	priority := uint8(5)

	// Market orders need immediate execution at current market price
	if order.OrderType() == domain.OrderTypeMarket {
		priority = 8
	}

	// Large orders require priority processing to minimize market impact
	orderValue := order.CalculateOrderValue()
	if orderValue > 100000 {
		priority = 7
	} else if orderValue > 10000 {
		priority = 6
	}

	// Stop loss orders are risk management tools requiring immediate attention
	if order.OrderType() == domain.OrderTypeStopLoss {
		priority = 7
	}

	return priority
}

type OrderStatusUpdate struct {
	OrderID         string               `json:"order_id"`
	UserID          string               `json:"user_id"`
	PreviousStatus  string               `json:"previous_status"`
	CurrentStatus   string               `json:"current_status"`
	UpdatedAt       time.Time            `json:"updated_at"`
	ExecutedAt      *time.Time           `json:"executed_at,omitempty"`
	ExecutionPrice  *float64             `json:"execution_price,omitempty"`
	MessageMetadata OrderMessageMetadata `json:"message_metadata"`
}

func (op *OrderProducer) GetQueueManager() *OrderQueueManager {
	return op.queueManager
}

func (op *OrderProducer) HealthCheck(ctx context.Context) error {
	return op.queueManager.HealthCheck(ctx)
}

func (op *OrderProducer) Close() error {
	// The underlying message handler will be closed by the container
	// This method exists for interface consistency and future extensibility
	return nil
}
