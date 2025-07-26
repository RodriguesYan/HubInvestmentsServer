package cache

import (
	"testing"
	"time"

	"HubInvestments/market_data/domain/model"
	"HubInvestments/shared/infra/cache"

	"github.com/stretchr/testify/assert"
)

// MockRepository mocks the database repository
type MockRepository struct {
	symbols []string
	data    []model.MarketDataModel
	err     error
}

func (m *MockRepository) GetMarketData(symbols []string) ([]model.MarketDataModel, error) {
	m.symbols = symbols
	return m.data, m.err
}

// MockCacheHandler mocks the cache handler
type MockCacheHandler struct {
	cache map[string]string
}

func NewMockCacheHandler() *MockCacheHandler {
	return &MockCacheHandler{
		cache: make(map[string]string),
	}
}

func (m *MockCacheHandler) Get(key string) (string, error) {
	value, exists := m.cache[key]
	if !exists {
		return "", cache.ErrCacheKeyNotFound
	}
	return value, nil
}

func (m *MockCacheHandler) Set(key string, value string, ttl time.Duration) error {
	m.cache[key] = value
	return nil
}

func (m *MockCacheHandler) Delete(key string) error {
	delete(m.cache, key)
	return nil
}

func TestMarketDataCacheRepository_GetMarketData_CacheHit(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()
	
	// Pre-populate cache
	mockCache.Set("market_data:AAPL", `{"Symbol":"AAPL","Name":"Apple Inc.","LastQuote":150.25,"Category":1}`, 5*time.Minute)
	
	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)
	
	// Act
	result, err := cacheRepo.GetMarketData([]string{"AAPL"})
	
	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "AAPL", result[0].Symbol)
	assert.Equal(t, float32(150.25), result[0].LastQuote)
	
	// Verify database was not called (cache hit)
	assert.Nil(t, mockRepo.symbols)
}

func TestMarketDataCacheRepository_GetMarketData_CacheMiss(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{
		data: []model.MarketDataModel{
			{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 150.25, Category: 1},
		},
		err: nil,
	}
	mockCache := NewMockCacheHandler()
	
	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)
	
	// Act
	result, err := cacheRepo.GetMarketData([]string{"AAPL"})
	
	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "AAPL", result[0].Symbol)
	assert.Equal(t, float32(150.25), result[0].LastQuote)
	
	// Verify database was called (cache miss)
	assert.Equal(t, []string{"AAPL"}, mockRepo.symbols)
	
	// Give async cache operation time to complete
	time.Sleep(50 * time.Millisecond)
	
	// Verify data was cached
	cachedValue, err := mockCache.Get("market_data:AAPL")
	assert.NoError(t, err)
	assert.Contains(t, cachedValue, "AAPL")
}

func TestMarketDataCacheRepository_GetMarketData_PartialCacheHit(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{
		data: []model.MarketDataModel{
			{Symbol: "GOOGL", Name: "Alphabet Inc.", LastQuote: 2750.50, Category: 1},
		},
		err: nil,
	}
	mockCache := NewMockCacheHandler()
	
	// Pre-populate cache with AAPL only
	mockCache.Set("market_data:AAPL", `{"Symbol":"AAPL","Name":"Apple Inc.","LastQuote":150.25,"Category":1}`, 5*time.Minute)
	
	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)
	
	// Act
	result, err := cacheRepo.GetMarketData([]string{"AAPL", "GOOGL"})
	
	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	
	// Verify both symbols are present
	symbols := make(map[string]bool)
	for _, item := range result {
		symbols[item.Symbol] = true
	}
	assert.True(t, symbols["AAPL"])
	assert.True(t, symbols["GOOGL"])
	
	// Verify only GOOGL was fetched from database
	assert.Equal(t, []string{"GOOGL"}, mockRepo.symbols)
}

func TestMarketDataCacheRepository_BuildCacheKey(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()
	
	cacheRepo := &MarketDataCacheRepository{
		dbRepo:      mockRepo,
		cacheClient: mockCache,
		ttl:         5 * time.Minute,
	}
	
	// Act & Assert
	assert.Equal(t, "market_data:AAPL", cacheRepo.buildCacheKey("AAPL"))
	assert.Equal(t, "market_data:AAPL", cacheRepo.buildCacheKey("aapl")) // Should uppercase
	assert.Equal(t, "market_data:GOOGL", cacheRepo.buildCacheKey("GOOGL"))
}

func TestMarketDataCacheRepository_InvalidateCache(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()
	
	// Pre-populate cache
	mockCache.Set("market_data:AAPL", "test_data", 5*time.Minute)
	mockCache.Set("market_data:GOOGL", "test_data", 5*time.Minute)
	
	cacheRepo := &MarketDataCacheRepository{
		dbRepo:      mockRepo,
		cacheClient: mockCache,
		ttl:         5 * time.Minute,
	}
	
	// Act
	err := cacheRepo.InvalidateCache([]string{"AAPL", "GOOGL"})
	
	// Assert
	assert.NoError(t, err)
	
	// Verify cache was cleared
	_, err1 := mockCache.Get("market_data:AAPL")
	_, err2 := mockCache.Get("market_data:GOOGL")
	assert.Equal(t, cache.ErrCacheKeyNotFound, err1)
	assert.Equal(t, cache.ErrCacheKeyNotFound, err2)
}
