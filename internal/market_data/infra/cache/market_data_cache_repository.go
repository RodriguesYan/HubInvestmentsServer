package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"HubInvestments/internal/market_data/domain/model"
	"HubInvestments/internal/market_data/domain/repository"
	"HubInvestments/shared/infra/cache"
)

// MarketDataCacheRepository implements cache-aside pattern for market data
// It decorates the existing database repository with Redis caching using your CacheHandler interface
type MarketDataCacheRepository struct {
	dbRepo      repository.IMarketDataRepository // Original database repository
	cacheClient cache.CacheHandler               // Your existing cache interface
	ttl         time.Duration                    // Time to live for cached data
}

// NewMarketDataCacheRepository creates a new cache repository that wraps the database repository
func NewMarketDataCacheRepository(
	dbRepo repository.IMarketDataRepository,
	cacheClient cache.CacheHandler,
	ttl time.Duration,
) repository.IMarketDataRepository {
	if ttl == 0 {
		ttl = 5 * time.Minute // Default 5 minutes TTL for market data
	}

	return &MarketDataCacheRepository{
		dbRepo:      dbRepo,
		cacheClient: cacheClient,
		ttl:         ttl,
	}
}

// GetMarketData implements cache-aside pattern
func (c *MarketDataCacheRepository) GetMarketData(symbols []string) ([]model.MarketDataModel, error) {
	// Step 1: Try to get data from cache first
	cachedData, missingSymbols := c.tryGetFromCache(symbols)

	// Step 2: If all data is in cache, return it
	if len(missingSymbols) == 0 {
		log.Printf("Cache HIT: All symbols found in cache: %v", symbols)
		return cachedData, nil
	}

	// Step 3: Fetch missing data from database
	log.Printf("Cache MISS: Fetching missing symbols from DB: %v", missingSymbols)
	dbData, err := c.dbRepo.GetMarketData(missingSymbols)
	if err != nil {
		// If DB fails, return cached data if we have any
		if len(cachedData) > 0 {
			log.Printf("DB error, returning partial cached data: %v", err)
			return cachedData, nil
		}
		return nil, fmt.Errorf("failed to fetch from database: %w", err)
	}

	// Step 4: Store new data in cache (fire and forget)
	go c.cacheNewData(dbData)

	// Step 5: Merge cached and database data
	allData := append(cachedData, dbData...)

	log.Printf("Cache-aside complete: returned %d items (cached: %d, db: %d)",
		len(allData), len(cachedData), len(dbData))

	return allData, nil
}

// tryGetFromCache attempts to retrieve data from cache for all symbols
func (c *MarketDataCacheRepository) tryGetFromCache(symbols []string) ([]model.MarketDataModel, []string) {
	var cachedData []model.MarketDataModel
	var missingSymbols []string

	for _, symbol := range symbols {
		cacheKey := c.buildCacheKey(symbol)

		cachedValue, err := c.cacheClient.Get(cacheKey)
		if err != nil {
			// Cache miss - add to missing list
			missingSymbols = append(missingSymbols, symbol)
			continue
		}

		// Cache hit - unmarshal the data
		var marketData model.MarketDataModel
		if err := json.Unmarshal([]byte(cachedValue), &marketData); err != nil {
			log.Printf("Failed to unmarshal cached data for %s: %v", symbol, err)
			missingSymbols = append(missingSymbols, symbol)
			continue
		}

		cachedData = append(cachedData, marketData)
	}

	return cachedData, missingSymbols
}

// cacheNewData stores the new data in cache asynchronously
func (c *MarketDataCacheRepository) cacheNewData(data []model.MarketDataModel) {
	for _, item := range data {
		cacheKey := c.buildCacheKey(item.Symbol)

		dataBytes, err := json.Marshal(item)
		if err != nil {
			log.Printf("Failed to marshal data for caching %s: %v", item.Symbol, err)
			continue
		}

		if err := c.cacheClient.Set(cacheKey, string(dataBytes), c.ttl); err != nil {
			log.Printf("Failed to cache data for %s: %v", item.Symbol, err)
		} else {
			log.Printf("Successfully cached data for %s", item.Symbol)
		}
	}
}

// buildCacheKey creates a consistent cache key for a symbol
func (c *MarketDataCacheRepository) buildCacheKey(symbol string) string {
	return fmt.Sprintf("market_data:%s", strings.ToUpper(symbol))
}

// InvalidateCache removes cached data for specific symbols
func (c *MarketDataCacheRepository) InvalidateCache(symbols []string) error {
	for _, symbol := range symbols {
		cacheKey := c.buildCacheKey(symbol)
		if err := c.cacheClient.Delete(cacheKey); err != nil {
			log.Printf("Failed to invalidate cache for %s: %v", symbol, err)
		} else {
			log.Printf("Successfully invalidated cache for %s", symbol)
		}
	}
	return nil
}

// WarmCache pre-loads frequently accessed symbols into cache
func (c *MarketDataCacheRepository) WarmCache(symbols []string) error {
	log.Printf("Warming cache for symbols: %v", symbols)

	// Fetch from database
	data, err := c.dbRepo.GetMarketData(symbols)
	if err != nil {
		return fmt.Errorf("failed to warm cache: %w", err)
	}

	// Store in cache
	c.cacheNewData(data)
	return nil
}
