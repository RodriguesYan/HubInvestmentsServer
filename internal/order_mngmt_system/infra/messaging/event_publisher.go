package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	msg "HubInvestments/shared/infra/messaging"
)

// TODO: implement opened and partial filled
type IEventPublisher interface {
	PublishOrderExecutedEvent(ctx context.Context, event *domain.OrderExecutedEvent) error
	PublishOrderFailedEvent(ctx context.Context, event *domain.OrderFailedEvent) error
	PublishOrderCancelledEvent(ctx context.Context, event *domain.OrderCancelledEvent) error
}

type EventPublisher struct {
	messageHandler msg.MessageHandler
	exchangeName   string
}

func NewEventPublisher(messageHandler msg.MessageHandler, exchangeName string) *EventPublisher {
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

	// Create position update message in the format expected by position worker
	positionUpdateMessage := map[string]interface{}{
		"order_id":              event.OrderID(),
		"user_id":               event.UserID(),
		"symbol":                event.Symbol,
		"order_side":            event.OrderSide.String(),
		"order_type":            event.OrderType.String(),
		"quantity":              event.Quantity,
		"execution_price":       event.ExecutionPrice,
		"executed_at":           event.ExecutedAt,
		"total_value":           event.TotalValue,
		"market_price_at_exec":  event.MarketPriceAtExec,
		"market_data_timestamp": event.MarketDataTimestamp,
		"message_metadata": map[string]interface{}{
			"message_id":       fmt.Sprintf("position_update_%s_%d", event.OrderID(), time.Now().UnixNano()),
			"correlation_id":   event.OrderID(),
			"timestamp":        time.Now(),
			"retry_attempt":    0,
			"priority":         uint8(1),
			"source":           "order_execution",
			"message_type":     "position_update",
			"processing_stage": "initial",
		},
	}

	messageBytes, err := json.Marshal(positionUpdateMessage)
	if err != nil {
		return fmt.Errorf("failed to serialize position update message: %w", err)
	}

	queueName := "positions.updates"
	messageID := fmt.Sprintf("position_update_%s_%d", event.OrderID(), time.Now().UnixNano())
	headers := map[string]interface{}{
		"event_type":     "OrderExecuted",
		"order_side":     event.OrderSide.String(),
		"symbol":         event.Symbol,
		"user_id":        event.UserID(),
		"execution_at":   event.ExecutedAt.Format(time.RFC3339),
		"message_id":     messageID,
		"correlation_id": event.OrderID(),
	}

	return p.publishEvent(ctx, queueName, messageBytes, messageID, headers)
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
	options := msg.PublishOptions{
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
