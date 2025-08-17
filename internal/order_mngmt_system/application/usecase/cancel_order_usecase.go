package usecase

import (
	"context"
	"fmt"
	"time"

	"HubInvestments/internal/order_mngmt_system/application/command"
	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/internal/order_mngmt_system/domain/repository"
)

// ICancelOrderUseCase defines the interface for cancelling orders
type ICancelOrderUseCase interface {
	Execute(ctx context.Context, cmd *command.CancelOrderCommand) (*command.CancelOrderResult, error)
}

// CancelOrderUseCase handles order cancellation with proper validation
type CancelOrderUseCase struct {
	orderRepository repository.IOrderRepository
}

// CancelOrderUseCaseConfig holds configuration for order cancellation
type CancelOrderUseCaseConfig struct {
	AllowCancellationGracePeriod time.Duration // Time after submission when cancellation is always allowed
	RequireReasonForCancellation bool          // Whether cancellation reason is mandatory
}

// NewCancelOrderUseCase creates a new cancel order use case
func NewCancelOrderUseCase(
	orderRepository repository.IOrderRepository,
) ICancelOrderUseCase {
	return &CancelOrderUseCase{
		orderRepository: orderRepository,
	}
}

// Execute processes the order cancellation request
func (uc *CancelOrderUseCase) Execute(ctx context.Context, cmd *command.CancelOrderCommand) (*command.CancelOrderResult, error) {
	// Step 1: Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cancellation command: %w", err)
	}

	// Step 2: Retrieve order from database
	order, err := uc.orderRepository.FindByID(cmd.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	// Step 3: Verify order belongs to the user
	if order.UserID() != cmd.UserID {
		return nil, fmt.Errorf("order not found") // Don't reveal that order exists for security
	}

	// Step 4: Validate order can be cancelled
	if err := uc.validateOrderCanBeCancelled(order); err != nil {
		return nil, fmt.Errorf("order cannot be cancelled: %w", err)
	}

	// Step 5: Cancel the order
	cancellationReason := string(cmd.GetReason())
	if err := uc.cancelOrder(ctx, order, cancellationReason); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	// Step 6: Create and return result
	result := &command.CancelOrderResult{
		OrderID:   order.ID(),
		Status:    string(order.Status()),
		Message:   fmt.Sprintf("Order %s has been cancelled successfully", order.ID()),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return result, nil
}

func (uc *CancelOrderUseCase) validateOrderCanBeCancelled(order *domain.Order) error {
	if !order.CanCancel() {
		return fmt.Errorf("order in status '%s' cannot be cancelled", order.Status())
	}

	switch order.Status() {
	case domain.OrderStatusPending:
		return nil

	case domain.OrderStatusProcessing:
		// Processing orders might be cancellable depending on timing
		// Check if order was recently submitted (grace period)
		gracePeriod := 30 * time.Second // 30 seconds grace period
		if time.Since(order.CreatedAt()) < gracePeriod {
			return nil
		}
		return fmt.Errorf("order is currently being processed and cannot be cancelled")

	case domain.OrderStatusExecuted:
		return fmt.Errorf("executed orders cannot be cancelled")

	case domain.OrderStatusFailed:
		return fmt.Errorf("failed orders are already in terminal state")

	case domain.OrderStatusCancelled:
		return fmt.Errorf("order is already cancelled")

	default:
		return fmt.Errorf("unknown order status: %s", order.Status())
	}
}

func (uc *CancelOrderUseCase) cancelOrder(ctx context.Context, order *domain.Order, reason string) error {
	if err := order.MarkAsCancelled(); err != nil {
		return fmt.Errorf("failed to mark order as cancelled: %w", err)
	}

	if err := uc.orderRepository.UpdateStatus(order.ID(), order.Status()); err != nil {
		return fmt.Errorf("failed to update order status in database: %w", err)
	}

	// Step 3: Integratiing in external vendor, we could:
	// - Notify external systems (broker, settlement, etc.)
	// - Release any reserved funds or positions
	// - Send notifications to the user
	// - Update related systems (risk management, reporting, etc.)

	return nil
}

func (uc *CancelOrderUseCase) CancelOrdersBatch(ctx context.Context, commands []*command.CancelOrderCommand) ([]*command.CancelOrderResult, []error) {
	results := make([]*command.CancelOrderResult, len(commands))
	errors := make([]error, len(commands))

	for i, cmd := range commands {
		result, err := uc.Execute(ctx, cmd)
		results[i] = result
		errors[i] = err
	}

	return results, errors
}

func (uc *CancelOrderUseCase) CancelOrdersByUser(ctx context.Context, userID, reason string) (*BatchCancellationResult, error) {
	orders, err := uc.orderRepository.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user orders: %w", err)
	}

	cancellableOrders := make([]*domain.Order, 0)
	for _, order := range orders {
		if order.CanCancel() {
			cancellableOrders = append(cancellableOrders, order)
		}
	}

	// Cancel each order
	result := &BatchCancellationResult{
		TotalOrders:     len(cancellableOrders),
		CancelledOrders: 0,
		FailedOrders:    0,
		Errors:          make([]string, 0),
	}

	for _, order := range cancellableOrders {
		cmd := &command.CancelOrderCommand{
			OrderID: order.ID(),
			UserID:  userID,
			Reason:  reason,
		}

		_, err := uc.Execute(ctx, cmd)
		if err != nil {
			result.FailedOrders++
			result.Errors = append(result.Errors, fmt.Sprintf("Order %s: %v", order.ID(), err))
		} else {
			result.CancelledOrders++
		}
	}

	return result, nil
}

func (uc *CancelOrderUseCase) CancelExpiredOrders(ctx context.Context, expirationTime time.Time) (*BatchCancellationResult, error) {
	// For now, this method is simplified
	return &BatchCancellationResult{
		TotalOrders:     0,
		CancelledOrders: 0,
		FailedOrders:    0,
		Errors:          []string{"CancelExpiredOrders not fully implemented yet"},
	}, nil
}

// BatchCancellationResult represents the result of batch cancellation operations
type BatchCancellationResult struct {
	TotalOrders     int      `json:"total_orders"`
	CancelledOrders int      `json:"cancelled_orders"`
	FailedOrders    int      `json:"failed_orders"`
	Errors          []string `json:"errors,omitempty"`
}

// GetSummary returns a human-readable summary of the batch operation
func (r *BatchCancellationResult) GetSummary() string {
	if r.TotalOrders == 0 {
		return "No orders found to cancel"
	}

	if r.FailedOrders == 0 {
		return fmt.Sprintf("Successfully cancelled %d out of %d orders", r.CancelledOrders, r.TotalOrders)
	}

	return fmt.Sprintf("Cancelled %d orders, failed to cancel %d orders out of %d total",
		r.CancelledOrders, r.FailedOrders, r.TotalOrders)
}
