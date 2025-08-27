package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	domain "HubInvestments/internal/order_mngmt_system/domain/model"
)

type IdempotencyKey struct {
	Key       string    `json:"key"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	OrderID   string    `json:"order_id,omitempty"`
	Status    string    `json:"status"`
	Result    string    `json:"result,omitempty"`
}

type IdempotencyStatus string

const (
	IdempotencyStatusPending   IdempotencyStatus = "PENDING"
	IdempotencyStatusCompleted IdempotencyStatus = "COMPLETED"
	IdempotencyStatusFailed    IdempotencyStatus = "FAILED"
	IdempotencyStatusExpired   IdempotencyStatus = "EXPIRED"
)

type IIdempotencyService interface {
	// GenerateKey creates an idempotency key based on order parameters
	GenerateKey(userID, symbol, orderType, orderSide string, quantity float64, price *float64) string

	// CheckIdempotency verifies if an operation has already been processed
	CheckIdempotency(ctx context.Context, key, userID string) (*IdempotencyResult, error)

	// StoreIdempotencyKey stores a new idempotency key with pending status
	StoreIdempotencyKey(ctx context.Context, key, userID string, ttl time.Duration) error

	// CompleteIdempotency marks an idempotency key as completed with result
	CompleteIdempotency(ctx context.Context, key, userID, orderID, result string) error

	// FailIdempotency marks an idempotency key as failed with error
	FailIdempotency(ctx context.Context, key, userID, errorMsg string) error

	// CleanupExpiredKeys removes expired idempotency keys
	CleanupExpiredKeys(ctx context.Context) error
}

type IdempotencyResult struct {
	IsProcessed bool              `json:"is_processed"`
	Status      IdempotencyStatus `json:"status"`
	OrderID     string            `json:"order_id,omitempty"`
	Result      string            `json:"result,omitempty"`
	Error       string            `json:"error,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
}

type IdempotencyService struct {
	repository IIdempotencyRepository
	defaultTTL time.Duration
}

type IIdempotencyRepository interface {
	Store(ctx context.Context, key *IdempotencyKey) error
	Get(ctx context.Context, key, userID string) (*IdempotencyKey, error)
	Update(ctx context.Context, key *IdempotencyKey) error
	Delete(ctx context.Context, key, userID string) error
	DeleteExpired(ctx context.Context) error
}

func NewIdempotencyService(repository IIdempotencyRepository) IIdempotencyService {
	return &IdempotencyService{
		repository: repository,
		defaultTTL: 24 * time.Hour, // Default TTL of 24 hours
	}
}

func (s *IdempotencyService) GenerateKey(userID, symbol, orderType, orderSide string, quantity float64, price *float64) string {
	var priceStr string
	if price != nil {
		priceStr = fmt.Sprintf("%.8f", *price)
	} else {
		priceStr = "MARKET"
	}

	// Combine all parameters that make an order unique
	data := fmt.Sprintf("%s:%s:%s:%s:%.8f:%s",
		userID, symbol, orderType, orderSide, quantity, priceStr)

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("order_%x", hash)
}

// CheckIdempotency verifies if an operation has already been processed
func (s *IdempotencyService) CheckIdempotency(ctx context.Context, key, userID string) (*IdempotencyResult, error) {
	idempotencyKey, err := s.repository.Get(ctx, key, userID)
	if err != nil {
		// Key not found means this is a new operation
		return &IdempotencyResult{
			IsProcessed: false,
			Status:      IdempotencyStatusPending,
		}, nil
	}

	// Check if key has expired
	if time.Now().After(idempotencyKey.ExpiresAt) {
		// Clean up expired key
		_ = s.repository.Delete(ctx, key, userID)
		return &IdempotencyResult{
			IsProcessed: false,
			Status:      IdempotencyStatusExpired,
		}, nil
	}

	// Key exists and is valid
	result := &IdempotencyResult{
		IsProcessed: true,
		Status:      IdempotencyStatus(idempotencyKey.Status),
		OrderID:     idempotencyKey.OrderID,
		Result:      idempotencyKey.Result,
		CreatedAt:   idempotencyKey.CreatedAt,
		ExpiresAt:   idempotencyKey.ExpiresAt,
	}

	return result, nil
}

// StoreIdempotencyKey stores a new idempotency key with pending status
func (s *IdempotencyService) StoreIdempotencyKey(ctx context.Context, key, userID string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = s.defaultTTL
	}

	now := time.Now()
	idempotencyKey := &IdempotencyKey{
		Key:       key,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		Status:    string(IdempotencyStatusPending),
	}

	return s.repository.Store(ctx, idempotencyKey)
}

func (s *IdempotencyService) CompleteIdempotency(ctx context.Context, key, userID, orderID, result string) error {
	idempotencyKey, err := s.repository.Get(ctx, key, userID)
	if err != nil {
		return fmt.Errorf("idempotency key not found: %w", err)
	}

	idempotencyKey.Status = string(IdempotencyStatusCompleted)
	idempotencyKey.OrderID = orderID
	idempotencyKey.Result = result

	return s.repository.Update(ctx, idempotencyKey)
}

func (s *IdempotencyService) FailIdempotency(ctx context.Context, key, userID, errorMsg string) error {
	idempotencyKey, err := s.repository.Get(ctx, key, userID)
	if err != nil {
		return fmt.Errorf("idempotency key not found: %w", err)
	}

	idempotencyKey.Status = string(IdempotencyStatusFailed)
	idempotencyKey.Result = errorMsg

	return s.repository.Update(ctx, idempotencyKey)
}

func (s *IdempotencyService) CleanupExpiredKeys(ctx context.Context) error {
	return s.repository.DeleteExpired(ctx)
}

// ValidateOrderForIdempotency validates if an order can be processed based on idempotency
func (s *IdempotencyService) ValidateOrderForIdempotency(ctx context.Context, order *domain.Order) (*IdempotencyResult, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Generate idempotency key for the order
	key := s.GenerateKey(
		order.UserID(),
		order.Symbol(),
		order.OrderType().String(),
		order.OrderSide().String(),
		order.Quantity(),
		order.Price(),
	)

	// Check if this order has already been processed
	return s.CheckIdempotency(ctx, key, order.UserID())
}

func (s *IdempotencyService) ProcessOrderWithIdempotency(
	ctx context.Context,
	order *domain.Order,
	processor func(context.Context, *domain.Order) (string, error),
) (*IdempotencyResult, error) {
	key := s.GenerateKey(
		order.UserID(),
		order.Symbol(),
		order.OrderType().String(),
		order.OrderSide().String(),
		order.Quantity(),
		order.Price(),
	)

	result, err := s.CheckIdempotency(ctx, key, order.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency: %w", err)
	}

	// If already processed, return existing result
	if result.IsProcessed {
		return result, nil
	}

	// Store pending idempotency key
	if err := s.StoreIdempotencyKey(ctx, key, order.UserID(), s.defaultTTL); err != nil {
		return nil, fmt.Errorf("failed to store idempotency key: %w", err)
	}

	orderID, err := processor(ctx, order)
	if err != nil {
		_ = s.FailIdempotency(ctx, key, order.UserID(), err.Error())
		return &IdempotencyResult{
			IsProcessed: true,
			Status:      IdempotencyStatusFailed,
			Error:       err.Error(),
		}, err
	}

	if err := s.CompleteIdempotency(ctx, key, order.UserID(), orderID, "Order submitted successfully"); err != nil {
		return nil, fmt.Errorf("failed to complete idempotency: %w", err)
	}

	return &IdempotencyResult{
		IsProcessed: true,
		Status:      IdempotencyStatusCompleted,
		OrderID:     orderID,
		Result:      "Order submitted successfully",
	}, nil
}
