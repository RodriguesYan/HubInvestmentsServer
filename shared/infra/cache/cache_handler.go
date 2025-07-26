package cache

import (
	"errors"
	"time"
)

var ErrCacheKeyNotFound = errors.New("cache key not found")

type CacheHandler interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Delete(key string) error
}
