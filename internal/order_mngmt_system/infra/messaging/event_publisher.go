package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/shared/infra/messaging"
)

// TODO: implement opened and partial filled
type IEventPublisher interface {
	PublishOrderExecutedEvent(ctx context.Context, event *domain.OrderExecutedEvent) error
	PublishOrderFailedEvent(ctx context.Context, event *domain.OrderFailedEvent) error
	PublishOrderCancelledEvent(ctx context.Context, event *domain.OrderCancelledEvent) error
}

// EventPublisher implements IEventPublisher using the messaging abstraction
type EventPublisher struct {
	messageHandler messaging.MessageHandler
	exchangeName   string
}

// NewEventPublisher creates a new event publisher using the messaging abstraction
func NewEventPublisher(messageHandler messaging.MessageHandler, exchangeName string) *EventPublisher {
	if exchangeName == "" {
		exchangeName = "orders.events"
	}
	return &EventPublisher{
		messageHandler: messageHandler,
		exchangeName:   exchangeName,
	}
}

type EventMessage struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	OccurredAt    time.Time              `json:"occurred_at"`
	EventData     map[string]interface{} `json:"event_data"`
	MessageID     string                 `json:"message_id"`
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Source        string                 `json:"source"`
}

func (p *EventPublisher) PublishOrderExecutedEvent(ctx context.Context, event *domain.OrderExecutedEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eventData := map[string]interface{}{
		"order_id":              event.OrderID(),
		"user_id":               event.UserID(),
		"symbol":                event.Symbol,
		"order_side":            string(event.OrderSide),
		"order_type":            string(event.OrderType),
		"quantity":              event.Quantity,
		"execution_price":       event.ExecutionPrice,
		"executed_at":           event.ExecutedAt,
		"total_value":           event.TotalValue,
		"market_price_at_exec":  event.MarketPriceAtExec,
		"market_data_timestamp": event.MarketDataTimestamp,
	}

	eventMessage := EventMessage{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		AggregateID:   event.AggregateID(),
		OccurredAt:    event.OccurredAt(),
		EventData:     eventData,
		MessageID:     fmt.Sprintf("event_%s_%d", event.EventID(), time.Now().UnixNano()),
		CorrelationID: event.OrderID(),
		Timestamp:     time.Now(),
		Source:        "order_execution",
	}

	messageBytes, err := json.Marshal(eventMessage)
	if err != nil {
		return fmt.Errorf("failed to serialize event message: %w", err)
	}

	queueName := "position.updates"
	headers := map[string]interface{}{
		"event_type":   event.EventType(),
		"order_side":   string(event.OrderSide),
		"symbol":       event.Symbol,
		"user_id":      event.UserID(),
		"execution_at": event.ExecutedAt.Format(time.RFC3339),
	}

	return p.publishEvent(ctx, queueName, messageBytes, eventMessage.MessageID, headers)
}

func (p *EventPublisher) PublishOrderFailedEvent(ctx context.Context, event *domain.OrderFailedEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eventData := map[string]interface{}{
		"order_id":       event.OrderID(),
		"user_id":        event.UserID(),
		"failure_reason": event.FailureReason,
		"failed_at":      event.FailedAt,
	}

	eventMessage := EventMessage{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		AggregateID:   event.AggregateID(),
		OccurredAt:    event.OccurredAt(),
		EventData:     eventData,
		MessageID:     fmt.Sprintf("event_%s_%d", event.EventID(), time.Now().UnixNano()),
		CorrelationID: event.OrderID(),
		Timestamp:     time.Now(),
		Source:        "order_processing",
	}

	messageBytes, err := json.Marshal(eventMessage)
	if err != nil {
		return fmt.Errorf("failed to serialize event message: %w", err)
	}

	queueName := "orders.failed"
	headers := map[string]interface{}{
		"event_type": event.EventType(),
		"user_id":    event.UserID(),
		"failed_at":  event.FailedAt.Format(time.RFC3339),
	}

	return p.publishEvent(ctx, queueName, messageBytes, eventMessage.MessageID, headers)
}

func (p *EventPublisher) PublishOrderCancelledEvent(ctx context.Context, event *domain.OrderCancelledEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eventData := map[string]interface{}{
		"order_id":      event.OrderID(),
		"user_id":       event.UserID(),
		"cancelled_at":  event.CancelledAt,
		"cancel_reason": event.CancelReason,
		"cancelled_by":  event.CancelledBy,
	}

	eventMessage := EventMessage{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		AggregateID:   event.AggregateID(),
		OccurredAt:    event.OccurredAt(),
		EventData:     eventData,
		MessageID:     fmt.Sprintf("event_%s_%d", event.EventID(), time.Now().UnixNano()),
		CorrelationID: event.OrderID(),
		Timestamp:     time.Now(),
		Source:        "order_cancellation",
	}

	messageBytes, err := json.Marshal(eventMessage)
	if err != nil {
		return fmt.Errorf("failed to serialize event message: %w", err)
	}

	queueName := "orders.cancelled"
	headers := map[string]interface{}{
		"event_type":   event.EventType(),
		"user_id":      event.UserID(),
		"cancelled_at": event.CancelledAt.Format(time.RFC3339),
		"cancelled_by": event.CancelledBy,
	}

	return p.publishEvent(ctx, queueName, messageBytes, eventMessage.MessageID, headers)
}

func (p *EventPublisher) publishEvent(
	ctx context.Context,
	queueName string,
	messageBytes []byte,
	messageID string,
	headers map[string]interface{},
) error {
	// For simple publish - using the queueName directly
	err := p.messageHandler.Publish(ctx, queueName, messageBytes)
	if err == nil {
		return nil
	}
	// Try with options if simple publish fails
	options := messaging.PublishOptions{
		QueueName:     queueName,
		Message:       messageBytes,
		MessageID:     messageID,
		Headers:       headers,
		Persistent:    true,
		CorrelationID: messageID,
	}

	err = p.messageHandler.PublishWithOptions(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to publish event to queue '%s': %w", queueName, err)
	}

	return nil

}
