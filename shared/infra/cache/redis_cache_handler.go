package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCacheHandler struct {
	redis *redis.Client
}

func NewRedisCacheHandler(redis *redis.Client) CacheHandler {
	return &RedisCacheHandler{redis: redis}
}

var ctx = context.Background()

func (r *RedisCacheHandler) Get(key string) (string, error) {
	val, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisCacheHandler) Set(key string, value string, ttl time.Duration) error {
	err := r.redis.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCacheHandler) Delete(key string) error {
	err := r.redis.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}
