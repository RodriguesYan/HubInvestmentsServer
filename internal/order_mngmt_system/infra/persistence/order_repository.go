package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
	"HubInvestments/internal/order_mngmt_system/domain/repository"
	"HubInvestments/internal/order_mngmt_system/infra/persistence/dto"
	"HubInvestments/shared/infra/database"

	"github.com/google/uuid"
)

type OrderRepository struct {
	db     database.Database
	mapper *dto.OrderMapper
}

func NewOrderRepository(db database.Database) repository.IOrderRepository {
	return &OrderRepository{
		db:     db,
		mapper: dto.NewOrderMapper(),
	}
}

func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}

	orderDTO, err := r.mapper.ToDTO(order)
	if err != nil {
		return fmt.Errorf("failed to convert order to DTO: %w", err)
	}

	query := `
		INSERT INTO orders (
			id, user_id, symbol, order_type, order_side, quantity, price, status,
			created_at, updated_at, executed_at, execution_price, 
			market_price_at_submission, market_data_timestamp, failure_reason,
			retry_count, processing_worker_id, external_order_id
		) VALUES (
			:id, :user_id, :symbol, :order_type, :order_side, :quantity, :price, :status,
			:created_at, :updated_at, :executed_at, :execution_price,
			:market_price_at_submission, :market_data_timestamp, :failure_reason,
			:retry_count, :processing_worker_id, :external_order_id
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at,
			executed_at = EXCLUDED.executed_at,
			execution_price = EXCLUDED.execution_price,
			failure_reason = EXCLUDED.failure_reason,
			retry_count = EXCLUDED.retry_count,
			processing_worker_id = EXCLUDED.processing_worker_id,
			external_order_id = EXCLUDED.external_order_id`

	_, err = r.db.ExecContext(ctx, query,
		orderDTO.ID, orderDTO.UserID, orderDTO.Symbol, orderDTO.OrderType, orderDTO.OrderSide,
		orderDTO.Quantity, orderDTO.Price, orderDTO.Status, orderDTO.CreatedAt, orderDTO.UpdatedAt,
		orderDTO.ExecutedAt, orderDTO.ExecutionPrice, orderDTO.MarketPriceAtSubmission,
		orderDTO.MarketDataTimestamp, orderDTO.FailureReason, orderDTO.RetryCount,
		orderDTO.ProcessingWorkerID, orderDTO.ExternalOrderID)

	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, orderID string) (*domain.Order, error) {
	var orderDTO dto.OrderDTO

	query := `
		SELECT id, user_id, symbol, order_type, order_side, quantity, price, status,
			   created_at, updated_at, executed_at, execution_price,
			   market_price_at_submission, market_data_timestamp, failure_reason,
			   retry_count, processing_worker_id, external_order_id
		FROM orders 
		WHERE id = $1`

	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID format: %w", err)
	}

	err = r.db.Get(&orderDTO, query, orderUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %s", orderID)
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	order, err := r.mapper.ToDomain(&orderDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTO to domain: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	var orderDTOs []*dto.OrderDTO

	query := `
		SELECT id, user_id, symbol, order_type, order_side, quantity, price, status,
			   created_at, updated_at, executed_at, execution_price,
			   market_price_at_submission, market_data_timestamp, failure_reason,
			   retry_count, processing_worker_id, external_order_id
		FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC`

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	err = r.db.Select(&orderDTOs, query, userIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by user ID: %w", err)
	}

	orders, err := r.mapper.ToOrderList(orderDTOs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTOs to domain: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) FindByUserIDAndStatus(ctx context.Context, userID string, status domain.OrderStatus) ([]*domain.Order, error) {
	var orderDTOs []*dto.OrderDTO

	query := `
		SELECT id, user_id, symbol, order_type, order_side, quantity, price, status,
			   created_at, updated_at, executed_at, execution_price,
			   market_price_at_submission, market_data_timestamp, failure_reason,
			   retry_count, processing_worker_id, external_order_id
		FROM orders 
		WHERE user_id = $1 AND status = $2 
		ORDER BY created_at DESC`

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	err = r.db.Select(&orderDTOs, query, userIDInt, status.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by user ID and status: %w", err)
	}

	orders, err := r.mapper.ToOrderList(orderDTOs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTOs to domain: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) FindByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	var orderDTOs []*dto.OrderDTO

	query := `
		SELECT id, user_id, symbol, order_type, order_side, quantity, price, status,
			   created_at, updated_at, executed_at, execution_price,
			   market_price_at_submission, market_data_timestamp, failure_reason,
			   retry_count, processing_worker_id, external_order_id
		FROM orders 
		WHERE status = $1 
		ORDER BY created_at DESC`

	err := r.db.Select(&orderDTOs, query, status.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by status: %w", err)
	}

	orders, err := r.mapper.ToOrderList(orderDTOs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTOs to domain: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	query := `
		UPDATE orders 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status.String(), orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", orderID)
	}

	return nil
}

func (r *OrderRepository) UpdateExecutionDetails(ctx context.Context, orderID string, executionPrice float64, executedAt time.Time) error {
	query := `
		UPDATE orders 
		SET execution_price = $1, 
			executed_at = $2, 
			status = $3,
			updated_at = CURRENT_TIMESTAMP 
		WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query, executionPrice, executedAt, domain.OrderStatusExecuted.String(), orderID)
	if err != nil {
		return fmt.Errorf("failed to update execution details: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", orderID)
	}

	return nil
}

func (r *OrderRepository) Delete(ctx context.Context, orderID string) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", orderID)
	}

	return nil
}

func (r *OrderRepository) CountOrdersByUserID(ctx context.Context, userID string) (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM orders WHERE user_id = $1`

	// Convert string userID to int for database query
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}

	err = r.db.Get(&count, query, userIDInt)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders by user ID: %w", err)
	}

	return count, nil
}

// FindOrderHistory retrieves paginated order history for a user
func (r *OrderRepository) FindOrderHistory(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	var orderDTOs []*dto.OrderDTO

	query := `
		SELECT id, user_id, symbol, order_type, order_side, quantity, price, status,
			   created_at, updated_at, executed_at, execution_price,
			   market_price_at_submission, market_data_timestamp, failure_reason,
			   retry_count, processing_worker_id, external_order_id
		FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	err = r.db.Select(&orderDTOs, query, userIDInt, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find order history: %w", err)
	}

	orders, err := r.mapper.ToOrderList(orderDTOs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTOs to domain: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) FindOrdersBySymbol(ctx context.Context, symbol string) ([]*domain.Order, error) {
	var orderDTOs []*dto.OrderDTO

	query := `
		SELECT id, user_id, symbol, order_type, order_side, quantity, price, status,
			   created_at, updated_at, executed_at, execution_price,
			   market_price_at_submission, market_data_timestamp, failure_reason,
			   retry_count, processing_worker_id, external_order_id
		FROM orders 
		WHERE symbol = $1 
		ORDER BY created_at DESC`

	err := r.db.Select(&orderDTOs, query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by symbol: %w", err)
	}

	orders, err := r.mapper.ToOrderList(orderDTOs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTOs to domain: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) FindOrdersByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*domain.Order, error) {
	var orderDTOs []*dto.OrderDTO

	query := `
		SELECT id, user_id, symbol, order_type, order_side, quantity, price, status,
			   created_at, updated_at, executed_at, execution_price,
			   market_price_at_submission, market_data_timestamp, failure_reason,
			   retry_count, processing_worker_id, external_order_id
		FROM orders 
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3 
		ORDER BY created_at DESC`

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	err = r.db.Select(&orderDTOs, query, userIDInt, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by date range: %w", err)
	}

	orders, err := r.mapper.ToOrderList(orderDTOs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTOs to domain: %w", err)
	}

	return orders, nil
}
