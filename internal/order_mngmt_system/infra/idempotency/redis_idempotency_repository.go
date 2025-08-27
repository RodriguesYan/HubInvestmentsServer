package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"HubInvestments/internal/order_mngmt_system/domain/service"
	"HubInvestments/shared/infra/cache"
)

type RedisIdempotencyRepository struct {
	cacheHandler cache.CacheHandler
	keyPrefix    string
}

func NewRedisIdempotencyRepository(cacheHandler cache.CacheHandler) service.IIdempotencyRepository {
	return &RedisIdempotencyRepository{
		cacheHandler: cacheHandler,
		keyPrefix:    "idempotency:",
	}
}

func (r *RedisIdempotencyRepository) Store(ctx context.Context, key *service.IdempotencyKey) error {
	if key == nil {
		return fmt.Errorf("idempotency key cannot be nil")
	}

	data, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal idempotency key: %w", err)
	}

	ttl := time.Until(key.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("idempotency key has already expired")
	}

	redisKey := r.buildRedisKey(key.Key, key.UserID)
	if err := r.cacheHandler.Set(redisKey, string(data), ttl); err != nil {
		return fmt.Errorf("failed to store idempotency key in Redis: %w", err)
	}

	return nil
}

func (r *RedisIdempotencyRepository) Get(ctx context.Context, key, userID string) (*service.IdempotencyKey, error) {
	redisKey := r.buildRedisKey(key, userID)

	data, err := r.cacheHandler.Get(redisKey)
	if err != nil {
		return nil, fmt.Errorf("idempotency key not found: %w", err)
	}

	var idempotencyKey service.IdempotencyKey
	if err := json.Unmarshal([]byte(data), &idempotencyKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal idempotency key: %w", err)
	}

	return &idempotencyKey, nil
}

func (r *RedisIdempotencyRepository) Update(ctx context.Context, key *service.IdempotencyKey) error {
	if key == nil {
		return fmt.Errorf("idempotency key cannot be nil")
	}

	// Check if key exists
	redisKey := r.buildRedisKey(key.Key, key.UserID)
	_, err := r.cacheHandler.Get(redisKey)
	if err != nil {
		return fmt.Errorf("idempotency key not found for update: %w", err)
	}

	data, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal updated idempotency key: %w", err)
	}

	ttl := time.Until(key.ExpiresAt)
	if ttl <= 0 {
		// Key has expired, delete it
		return r.Delete(ctx, key.Key, key.UserID)
	}

	if err := r.cacheHandler.Set(redisKey, string(data), ttl); err != nil {
		return fmt.Errorf("failed to update idempotency key in Redis: %w", err)
	}

	return nil
}

func (r *RedisIdempotencyRepository) Delete(ctx context.Context, key, userID string) error {
	redisKey := r.buildRedisKey(key, userID)

	if err := r.cacheHandler.Delete(redisKey); err != nil {
		return fmt.Errorf("failed to delete idempotency key from Redis: %w", err)
	}

	return nil
}

// DeleteExpired removes expired idempotency keys
// Note: Redis automatically handles TTL expiration, so this is mainly for cleanup
func (r *RedisIdempotencyRepository) DeleteExpired(ctx context.Context) error {
	// Redis automatically expires keys based on TTL
	// This method is implemented for interface compliance
	// In a production system, you might want to scan for expired keys and clean them up
	return nil
}

func (r *RedisIdempotencyRepository) buildRedisKey(key, userID string) string {
	return fmt.Sprintf("%s%s:%s", r.keyPrefix, userID, key)
}

func (r *RedisIdempotencyRepository) HealthCheck(ctx context.Context) error {
	testKey := r.buildRedisKey("health_check", "test")
	testValue := "ping"
	testTTL := 10 * time.Second

	if err := r.cacheHandler.Set(testKey, testValue, testTTL); err != nil {
		return fmt.Errorf("Redis health check failed on SET: %w", err)
	}

	value, err := r.cacheHandler.Get(testKey)
	if err != nil {
		return fmt.Errorf("Redis health check failed on GET: %w", err)
	}

	if value != testValue {
		return fmt.Errorf("Redis health check failed: expected %s, got %s", testValue, value)
	}

	_ = r.cacheHandler.Delete(testKey)

	return nil
}

// GetStats returns statistics about idempotency keys (if supported by Redis)
func (r *RedisIdempotencyRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// This is a basic implementation
	// In a production system, you might want to scan Redis for keys matching the prefix
	// and provide more detailed statistics
	stats["repository_type"] = "redis"
	stats["key_prefix"] = r.keyPrefix
	stats["status"] = "active"

	return stats, nil
}

type MockIdempotencyRepository struct {
	keys map[string]*service.IdempotencyKey
}

func NewMockIdempotencyRepository() service.IIdempotencyRepository {
	return &MockIdempotencyRepository{
		keys: make(map[string]*service.IdempotencyKey),
	}
}

func (m *MockIdempotencyRepository) Store(ctx context.Context, key *service.IdempotencyKey) error {
	if key == nil {
		return fmt.Errorf("idempotency key cannot be nil")
	}

	mockKey := fmt.Sprintf("%s:%s", key.UserID, key.Key)
	m.keys[mockKey] = key
	return nil
}

func (m *MockIdempotencyRepository) Get(ctx context.Context, key, userID string) (*service.IdempotencyKey, error) {
	mockKey := fmt.Sprintf("%s:%s", userID, key)
	idempotencyKey, exists := m.keys[mockKey]
	if !exists {
		return nil, fmt.Errorf("idempotency key not found")
	}

	// Check expiration
	if time.Now().After(idempotencyKey.ExpiresAt) {
		delete(m.keys, mockKey)
		return nil, fmt.Errorf("idempotency key expired")
	}

	return idempotencyKey, nil
}

func (m *MockIdempotencyRepository) Update(ctx context.Context, key *service.IdempotencyKey) error {
	if key == nil {
		return fmt.Errorf("idempotency key cannot be nil")
	}

	mockKey := fmt.Sprintf("%s:%s", key.UserID, key.Key)
	if _, exists := m.keys[mockKey]; !exists {
		return fmt.Errorf("idempotency key not found for update")
	}

	m.keys[mockKey] = key
	return nil
}

func (m *MockIdempotencyRepository) Delete(ctx context.Context, key, userID string) error {
	mockKey := fmt.Sprintf("%s:%s", userID, key)
	delete(m.keys, mockKey)
	return nil
}

func (m *MockIdempotencyRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for key, idempotencyKey := range m.keys {
		if now.After(idempotencyKey.ExpiresAt) {
			delete(m.keys, key)
		}
	}
	return nil
}
