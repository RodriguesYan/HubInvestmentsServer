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
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	val, err := rdb.Get(ctx, "key").Result()

	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisCacheHandler) Set(key string, value string, ttl time.Duration) error {
	return nil //TODO: fazer implementaçao do set aqui
}

func (r *RedisCacheHandler) Delete(key string) error {
	return nil
}
