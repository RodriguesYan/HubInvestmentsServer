package cache

import "HubInvestments/internal/market_data/domain/repository"

// CacheManager provides cache management operations
type CacheManager interface {
	InvalidateCache(symbols []string) error
	WarmCache(symbols []string) error
}

// GetCacheManager extracts cache management interface from repository
func GetCacheManager(repo repository.IMarketDataRepository) CacheManager {
	if cacheRepo, ok := repo.(*MarketDataCacheRepository); ok {
		return cacheRepo
	}
	// If not a cache repository, return no-op implementation
	return &noCacheManager{}
}

// noCacheManager is a no-op implementation for when cache is not used
type noCacheManager struct{}

func (n *noCacheManager) InvalidateCache(symbols []string) error {
	// No-op: cache not available
	return nil
}

func (n *noCacheManager) WarmCache(symbols []string) error {
	// No-op: cache not available
	return nil
}
