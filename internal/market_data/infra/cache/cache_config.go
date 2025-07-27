package cache

import "time"

// CacheConfig holds configuration for market data caching
type CacheConfig struct {
	TTL           time.Duration
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:           5 * time.Minute,
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
	}
}

// MarketDataCacheTTL returns recommended TTL for different market data types
func MarketDataCacheTTL() map[string]time.Duration {
	return map[string]time.Duration{
		"company_info":    1 * time.Hour,  // Longer for static data
		"historical_data": 24 * time.Hour, // Very long for historical data
	}
}
