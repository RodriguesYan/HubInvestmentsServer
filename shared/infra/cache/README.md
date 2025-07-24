# Cache Abstraction Layer

This package provides a cache abstraction layer that encapsulates caching operations, making it easy to switch between different cache implementations (Redis, Memcached, In-Memory, etc.) in the future without changing application code.

## Problem Solved

Previously, applications would directly depend on specific cache implementations:

```go
// Old approach - direct Redis dependency
import "github.com/redis/go-redis/v9"

type MarketDataService struct {
    redis *redis.Client  // Direct dependency on Redis
}

func (s *MarketDataService) GetMarketData(symbol string) (*MarketData, error) {
    // Direct Redis operations scattered throughout the code
    val, err := s.redis.Get(ctx, "market_data:"+symbol).Result()
    if err == redis.Nil {
        // Cache miss - fetch from database
        data := s.fetchFromDatabase(symbol)
        s.redis.Set(ctx, "market_data:"+symbol, data, 5*time.Minute)
        return data, nil
    }
    // Parse and return cached data...
}
```

If you wanted to switch from Redis to Memcached or add an in-memory fallback, you would need to:
1. Change all service implementations that use caching
2. Update dependency injection containers
3. Modify every place where cache operations are performed
4. Handle different API semantics between cache implementations
5. Potentially break existing functionality

## Solution

The cache abstraction layer provides:

1. **Generic Cache Interface**: All services depend on the `CacheHandler` interface, not Redis directly
2. **Single Point of Change**: To switch cache implementations, you only need to change the connection factory
3. **Future-Proof**: Easy to add support for Memcached, In-Memory caches, or distributed cache solutions
4. **Consistent API**: Uniform interface regardless of underlying cache technology
5. **Cache-Aside Pattern Support**: Built-in support for the most common caching pattern

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Service       │    │   CacheHandler   │    │ Redis           │
│                 │───▶│   Interface      │◄───│ Implementation  │
│ (MarketData,    │    │                  │    │                 │
│  Portfolio,     │    │ - Get()          │    │ (Currently)     │
│  Watchlist)     │    │ - Set()          │    └─────────────────┘
└─────────────────┘    │ - Delete()       │    ┌─────────────────┐
                       └──────────────────┘    │ Memcached       │
                                              │ Implementation  │
                                              │                 │
                                              │ (Future)        │
                                              └─────────────────┘
```

## Current Implementation

### Redis Implementation
- **File**: `redis_cache_handler.go`
- **Benefits**: High performance, persistence options, advanced data structures
- **Status**: Current implementation with basic Get/Set/Delete operations
- **Connection**: Uses `github.com/redis/go-redis/v9` client

## Interface Definition

```go
type CacheHandler interface {
    Get(key string) (string, error)
    Set(key string, value string, ttl time.Duration) error
    Delete(key string) error
}
```

### Method Descriptions

- **Get(key)**: Retrieves a value from cache by key. Returns empty string and error if key doesn't exist
- **Set(key, value, ttl)**: Stores a value in cache with specified Time-To-Live duration
- **Delete(key)**: Removes a key from cache. Returns error if operation fails

## Usage

### Basic Usage Pattern (Cache-Aside)

```go
package service

import (
    "time"
    "HubInvestments/shared/infra/cache"
)

type MarketDataService struct {
    cache cache.CacheHandler
    db    database.Database
}

func NewMarketDataService(cache cache.CacheHandler, db database.Database) *MarketDataService {
    return &MarketDataService{cache: cache, db: db}
}

func (s *MarketDataService) GetMarketData(symbol string) (*MarketData, error) {
    cacheKey := "market_data:" + symbol
    
    // 1. Try to get from cache first
    cachedData, err := s.cache.Get(cacheKey)
    if err == nil && cachedData != "" {
        // Cache hit - deserialize and return
        var data MarketData
        json.Unmarshal([]byte(cachedData), &data)
        return &data, nil
    }
    
    // 2. Cache miss - fetch from database
    data, err := s.fetchFromDatabase(symbol)
    if err != nil {
        return nil, err
    }
    
    // 3. Store in cache for future requests
    serializedData, _ := json.Marshal(data)
    s.cache.Set(cacheKey, string(serializedData), 5*time.Minute)
    
    return data, nil
}
```

### Dependency Injection Setup

```go
package main

import (
    "github.com/redis/go-redis/v9"
    "HubInvestments/shared/infra/cache"
)

func main() {
    // Create Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
    
    // Create cache handler
    cacheHandler := cache.NewRedisCacheHandler(redisClient)
    
    // Inject into services
    marketDataService := NewMarketDataService(cacheHandler, db)
}
```

## Cache Key Strategies

### Recommended Naming Conventions

```go
// Market Data
"market_data:{symbol}"                    // Individual symbols
"market_data:bulk:{hash}"                 // Multiple symbols (hash of symbol list)

// Portfolio Data
"portfolio:{user_id}"                     // User portfolio summary
"portfolio:positions:{user_id}"           // User positions only
"portfolio:balance:{user_id}"             // User balance only

// Watchlists
"watchlist:{user_id}:{watchlist_id}"      // Individual watchlist
"watchlists:{user_id}"                    // All user watchlists

// User Sessions
"session:{token_hash}"                    // Authentication sessions
"user:{user_id}:permissions"              // User permissions cache
```

### TTL (Time-To-Live) Recommendations

```go
const (
    // Real-time market data
    MarketDataTTL = 30 * time.Second
    
    // Portfolio data (changes less frequently)
    PortfolioTTL = 5 * time.Minute
    
    // User sessions
    SessionTTL = 24 * time.Hour
    
    // Watchlists (relatively static)
    WatchlistTTL = 1 * time.Hour
    
    // Reference data (symbols, instruments)
    ReferenceDataTTL = 24 * time.Hour
)
```

## Error Handling

### Cache Miss vs Cache Error

```go
func (s *Service) GetData(key string) (*Data, error) {
    value, err := s.cache.Get(key)
    
    if err != nil {
        // Cache error - log but don't fail the request
        log.Printf("Cache error for key %s: %v", key, err)
        // Fall back to database
        return s.fetchFromDatabase(key)
    }
    
    if value == "" {
        // Cache miss - normal case
        return s.fetchFromDatabase(key)
    }
    
    // Cache hit
    return s.parseValue(value), nil
}
```

### Graceful Degradation

```go
func (s *Service) SetCache(key, value string, ttl time.Duration) {
    err := s.cache.Set(key, value, ttl)
    if err != nil {
        // Log cache errors but don't fail the operation
        log.Printf("Failed to cache key %s: %v", key, err)
        // Application continues to work without caching
    }
}
```

## Testing

### Mock Implementation for Testing

```go
package cache

import "time"

type MockCacheHandler struct {
    data map[string]string
}

func NewMockCacheHandler() CacheHandler {
    return &MockCacheHandler{
        data: make(map[string]string),
    }
}

func (m *MockCacheHandler) Get(key string) (string, error) {
    value, exists := m.data[key]
    if !exists {
        return "", errors.New("key not found")
    }
    return value, nil
}

func (m *MockCacheHandler) Set(key string, value string, ttl time.Duration) error {
    m.data[key] = value
    return nil
}

func (m *MockCacheHandler) Delete(key string) error {
    delete(m.data, key)
    return nil
}
```

### Unit Test Example

```go
func TestMarketDataService_GetMarketData(t *testing.T) {
    // Arrange
    mockCache := cache.NewMockCacheHandler()
    service := NewMarketDataService(mockCache, mockDB)
    
    // Act
    data, err := service.GetMarketData("AAPL")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, data)
    
    // Verify data was cached
    cached, _ := mockCache.Get("market_data:AAPL")
    assert.NotEmpty(t, cached)
}
```

## Future Extensibility

### Adding Memcached Support

When you need to add Memcached support, you would:

1. **Create the implementation file** (`memcached_cache_handler.go`):
```go
package cache

import (
    "time"
    "github.com/bradfitz/gomemcache/memcache"
)

type MemcachedCacheHandler struct {
    client *memcache.Client
}

func NewMemcachedCacheHandler(client *memcache.Client) CacheHandler {
    return &MemcachedCacheHandler{client: client}
}

func (m *MemcachedCacheHandler) Get(key string) (string, error) {
    item, err := m.client.Get(key)
    if err != nil {
        return "", err
    }
    return string(item.Value), nil
}

func (m *MemcachedCacheHandler) Set(key string, value string, ttl time.Duration) error {
    item := &memcache.Item{
        Key:        key,
        Value:      []byte(value),
        Expiration: int32(ttl.Seconds()),
    }
    return m.client.Set(item)
}

func (m *MemcachedCacheHandler) Delete(key string) error {
    return m.client.Delete(key)
}
```

2. **Update dependency injection** to choose between implementations:
```go
func CreateCacheHandler(cacheType string) CacheHandler {
    switch cacheType {
    case "redis":
        return NewRedisCacheHandler(redisClient)
    case "memcached":
        return NewMemcachedCacheHandler(memcachedClient)
    default:
        return NewInMemoryCacheHandler()
    }
}
```

3. **All existing services continue to work** without any changes!

### In-Memory Cache for Development

```go
type InMemoryCacheHandler struct {
    data map[string]cacheItem
    mu   sync.RWMutex
}

type cacheItem struct {
    value     string
    expiresAt time.Time
}
```

## Performance Considerations

### Connection Pooling
- Redis client automatically handles connection pooling
- Configure pool size based on expected concurrency:
```go
redisClient := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     100,  // Adjust based on load
    MinIdleConns: 10,
})
```

### Serialization
- Use efficient serialization for complex objects (JSON, MessagePack, Protobuf)
- Consider compression for large objects
- Cache serialized strings to avoid repeated serialization

### Monitoring
- Track cache hit/miss ratios
- Monitor cache operation latency
- Set up alerts for cache failures

## Security Considerations

### Redis Security
- Use authentication: `AUTH` command
- Enable TLS for production deployments
- Restrict network access with firewalls
- Regular security updates

### Key Security
- Avoid storing sensitive data in cache keys
- Use hash functions for user-identifiable keys
- Implement proper key expiration

## Current Status & TODOs

### ✅ Completed
- [x] Basic cache interface definition
- [x] Redis implementation structure
- [x] Connection factory pattern ready

### ⏳ In Progress / Needs Implementation
- [ ] Complete Redis Set() method implementation
- [ ] Complete Redis Delete() method implementation
- [ ] Fix Redis client reuse (currently creates new client per operation)
- [ ] Add proper error handling and logging
- [ ] Add connection pooling configuration
- [ ] Add health checks and monitoring

### 🔮 Future Enhancements
- [ ] Memcached implementation
- [ ] In-memory cache for development/testing
- [ ] Distributed cache support
- [ ] Cache warming strategies
- [ ] Advanced TTL management
- [ ] Cache invalidation patterns
- [ ] Metrics and monitoring integration

## File Structure

```
shared/infra/cache/
├── README.md                    # This documentation
├── cache_handler.go             # Interface definitions
├── redis_cache_handler.go       # Redis implementation
├── mock_cache_handler.go        # Mock for testing (future)
└── in_memory_cache_handler.go   # In-memory implementation (future)
```

## Integration Examples

Once implemented, this cache layer will be used in:
- **Market Data Service**: Cache market quotes and symbol information
- **Portfolio Service**: Cache portfolio summaries and calculations
- **Watchlist Service**: Cache user watchlists for fast access
- **Authentication Service**: Cache user sessions and permissions

## Benefits

1. **Technology Independence**: Services don't depend on specific cache implementations
2. **Easy Testing**: Inject mock implementations for unit tests
3. **Performance**: Strategic caching reduces database load
4. **Scalability**: Cache layer helps handle high concurrent loads
5. **Flexibility**: Easy to switch between cache technologies
6. **Future-Proof**: Add new cache implementations without breaking existing code
7. **Consistency**: Uniform caching API across all services

## Summary

You now have a solid foundation for a cache abstraction layer that:
- ✅ **Provides a clean interface** for caching operations
- ✅ **Supports the cache-aside pattern** for optimal performance
- ✅ **Is technology-agnostic** and future-proof
- ⏳ **Redis implementation in progress** with basic structure
- 🔮 **Ready for extension** with additional cache technologies

When the Redis implementation is completed, you'll have a production-ready caching solution that can significantly improve application performance! 