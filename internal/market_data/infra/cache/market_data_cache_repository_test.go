package cache

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"HubInvestments/internal/market_data/domain/model"
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
	cache       map[string]string
	getError    error
	setError    error
	deleteError error
}

func NewMockCacheHandler() *MockCacheHandler {
	return &MockCacheHandler{
		cache: make(map[string]string),
	}
}

func (m *MockCacheHandler) Get(key string) (string, error) {
	if m.getError != nil {
		return "", m.getError
	}
	value, exists := m.cache[key]
	if !exists {
		return "", cache.ErrCacheKeyNotFound
	}
	return value, nil
}

func (m *MockCacheHandler) Set(key string, value string, ttl time.Duration) error {
	if m.setError != nil {
		return m.setError
	}
	m.cache[key] = value
	return nil
}

func (m *MockCacheHandler) Delete(key string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.cache, key)
	return nil
}

func TestNewMarketDataCacheRepository(t *testing.T) {
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()

	t.Run("with default TTL", func(t *testing.T) {
		cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 0)
		assert.NotNil(t, cacheRepo)
	})

	t.Run("with custom TTL", func(t *testing.T) {
		customTTL := 10 * time.Minute
		cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, customTTL)
		assert.NotNil(t, cacheRepo)
	})
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

func TestMarketDataCacheRepository_GetMarketData_DatabaseError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{
		data: nil,
		err:  errors.New("database connection failed"),
	}
	mockCache := NewMockCacheHandler()

	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

	// Act
	result, err := cacheRepo.GetMarketData([]string{"AAPL"})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch from database")
	assert.Nil(t, result)
}

func TestMarketDataCacheRepository_GetMarketData_DatabaseErrorWithPartialCache(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{
		data: nil,
		err:  errors.New("database connection failed"),
	}
	mockCache := NewMockCacheHandler()

	// Pre-populate cache with one symbol
	mockCache.Set("market_data:AAPL", `{"Symbol":"AAPL","Name":"Apple Inc.","LastQuote":150.25,"Category":1}`, 5*time.Minute)

	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

	// Act - Request AAPL (cached) and GOOGL (not cached, will fail)
	result, err := cacheRepo.GetMarketData([]string{"AAPL", "GOOGL"})

	// Assert - Should return cached data when DB fails
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "AAPL", result[0].Symbol)
	assert.Equal(t, []string{"GOOGL"}, mockRepo.symbols)
}

func TestMarketDataCacheRepository_GetMarketData_CacheError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{
		data: []model.MarketDataModel{
			{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 150.25, Category: 1},
		},
		err: nil,
	}
	mockCache := NewMockCacheHandler()
	mockCache.getError = errors.New("cache connection failed")

	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

	// Act
	result, err := cacheRepo.GetMarketData([]string{"AAPL"})

	// Assert - Should fall back to database when cache fails
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "AAPL", result[0].Symbol)
	assert.Equal(t, []string{"AAPL"}, mockRepo.symbols)
}

func TestMarketDataCacheRepository_GetMarketData_InvalidCachedJSON(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{
		data: []model.MarketDataModel{
			{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 150.25, Category: 1},
		},
		err: nil,
	}
	mockCache := NewMockCacheHandler()

	// Pre-populate cache with invalid JSON
	mockCache.Set("market_data:AAPL", `{"Symbol":"AAPL","Name":"Apple Inc.","LastQuote":invalid}`, 5*time.Minute)

	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

	// Act
	result, err := cacheRepo.GetMarketData([]string{"AAPL"})

	// Assert - Should fall back to database when cached JSON is invalid
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "AAPL", result[0].Symbol)
	assert.Equal(t, []string{"AAPL"}, mockRepo.symbols)
}

func TestMarketDataCacheRepository_GetMarketData_EmptySymbols(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()

	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

	// Act
	result, err := cacheRepo.GetMarketData([]string{})

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Nil(t, mockRepo.symbols)
}

func TestMarketDataCacheRepository_GetMarketData_NilSymbols(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()

	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

	// Act
	result, err := cacheRepo.GetMarketData(nil)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestMarketDataCacheRepository_GetMarketData_LargeBatch(t *testing.T) {
	// Arrange
	symbols := make([]string, 100)
	expectedData := make([]model.MarketDataModel, 100)
	for i := 0; i < 100; i++ {
		symbol := fmt.Sprintf("SYM%03d", i)
		symbols[i] = symbol
		expectedData[i] = model.MarketDataModel{
			Symbol:    symbol,
			Name:      fmt.Sprintf("Company %d", i),
			LastQuote: float32(100 + i),
			Category:  1,
		}
	}

	mockRepo := &MockRepository{
		data: expectedData,
		err:  nil,
	}
	mockCache := NewMockCacheHandler()

	cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

	// Act
	result, err := cacheRepo.GetMarketData(symbols)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 100)
	assert.Equal(t, symbols, mockRepo.symbols)
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
	assert.Equal(t, "market_data:", cacheRepo.buildCacheKey(""))   // Empty symbol
	assert.Equal(t, "market_data: ", cacheRepo.buildCacheKey(" ")) // Space
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

func TestMarketDataCacheRepository_InvalidateCache_WithErrors(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()
	mockCache.deleteError = errors.New("cache delete failed")

	cacheRepo := &MarketDataCacheRepository{
		dbRepo:      mockRepo,
		cacheClient: mockCache,
		ttl:         5 * time.Minute,
	}

	// Act
	err := cacheRepo.InvalidateCache([]string{"AAPL"})

	// Assert - Should not return error, just log it
	assert.NoError(t, err)
}

func TestMarketDataCacheRepository_WarmCache(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{
		data: []model.MarketDataModel{
			{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 150.25, Category: 1},
			{Symbol: "GOOGL", Name: "Alphabet Inc.", LastQuote: 2750.50, Category: 1},
		},
		err: nil,
	}
	mockCache := NewMockCacheHandler()

	cacheRepo := &MarketDataCacheRepository{
		dbRepo:      mockRepo,
		cacheClient: mockCache,
		ttl:         5 * time.Minute,
	}

	// Act
	err := cacheRepo.WarmCache([]string{"AAPL", "GOOGL"})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, []string{"AAPL", "GOOGL"}, mockRepo.symbols)

	// Give async cache operation time to complete
	time.Sleep(50 * time.Millisecond)

	// Verify data was cached
	cachedAAPL, err1 := mockCache.Get("market_data:AAPL")
	cachedGOOGL, err2 := mockCache.Get("market_data:GOOGL")
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Contains(t, cachedAAPL, "AAPL")
	assert.Contains(t, cachedGOOGL, "GOOGL")
}

func TestMarketDataCacheRepository_WarmCache_DatabaseError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{
		data: nil,
		err:  errors.New("database error"),
	}
	mockCache := NewMockCacheHandler()

	cacheRepo := &MarketDataCacheRepository{
		dbRepo:      mockRepo,
		cacheClient: mockCache,
		ttl:         5 * time.Minute,
	}

	// Act
	err := cacheRepo.WarmCache([]string{"AAPL"})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to warm cache")
}

func TestMarketDataCacheRepository_CacheNewData_SetError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()
	mockCache.setError = errors.New("cache set failed")

	cacheRepo := &MarketDataCacheRepository{
		dbRepo:      mockRepo,
		cacheClient: mockCache,
		ttl:         5 * time.Minute,
	}

	testData := []model.MarketDataModel{
		{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 150.25, Category: 1},
	}

	// Act - This should not panic or return error (it logs internally)
	cacheRepo.cacheNewData(testData)

	// Assert - Just verify it doesn't crash
	assert.True(t, true) // Test passes if no panic occurs
}

func TestMarketDataCacheRepository_CacheNewData_MarshalError(t *testing.T) {
	// This test is for completeness, though it's hard to trigger marshal errors
	// with valid MarketDataModel structs in Go
	mockRepo := &MockRepository{}
	mockCache := NewMockCacheHandler()

	cacheRepo := &MarketDataCacheRepository{
		dbRepo:      mockRepo,
		cacheClient: mockCache,
		ttl:         5 * time.Minute,
	}

	// Act with empty data
	cacheRepo.cacheNewData([]model.MarketDataModel{})

	// Assert - Should handle gracefully
	assert.True(t, true)
}

func TestMarketDataCacheRepository_EdgeCases(t *testing.T) {
	t.Run("symbol case sensitivity", func(t *testing.T) {
		mockRepo := &MockRepository{
			data: []model.MarketDataModel{
				{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 150.25, Category: 1},
			},
		}
		mockCache := NewMockCacheHandler()

		cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

		// Request with lowercase
		result, err := cacheRepo.GetMarketData([]string{"aapl"})
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		// Cache key should be uppercase
		time.Sleep(50 * time.Millisecond)
		cachedValue, err := mockCache.Get("market_data:AAPL")
		assert.NoError(t, err)
		assert.Contains(t, cachedValue, "AAPL")
	})

	t.Run("duplicate symbols in request", func(t *testing.T) {
		mockRepo := &MockRepository{
			data: []model.MarketDataModel{
				{Symbol: "AAPL", Name: "Apple Inc.", LastQuote: 150.25, Category: 1},
			},
		}
		mockCache := NewMockCacheHandler()

		cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

		// Request with duplicate symbols
		result, err := cacheRepo.GetMarketData([]string{"AAPL", "AAPL", "aapl"})
		assert.NoError(t, err)
		// Should handle duplicates gracefully (implementation dependent)
		assert.NotEmpty(t, result)
	})

	t.Run("very long symbol names", func(t *testing.T) {
		longSymbol := strings.Repeat("A", 100)
		mockRepo := &MockRepository{
			data: []model.MarketDataModel{
				{Symbol: longSymbol, Name: "Long Symbol Corp", LastQuote: 100.0, Category: 1},
			},
		}
		mockCache := NewMockCacheHandler()

		cacheRepo := NewMarketDataCacheRepository(mockRepo, mockCache, 5*time.Minute)

		result, err := cacheRepo.GetMarketData([]string{longSymbol})
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, longSymbol, result[0].Symbol)
	})
}
