package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

type Cache interface {
	Key(exchange, address string) string
	Store(ctx context.Context, key string, data interface{}, ttl time.Duration) error
	Value(ctx context.Context, key string) (interface{}, error)
}

type redisCache struct {
	redis *redis.Client
	log   *zap.Logger
}

func (r *redisCache) Store(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		r.log.Error(err.Error())
		return err
	}

	if err := r.redis.Set(ctx, key, b, ttl).Err(); err != nil {
		r.log.Error(err.Error())
		return err
	}

	return nil
}

func (r *redisCache) Value(ctx context.Context, key string) (interface{}, error) {
	val, err := r.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		r.log.Error(err.Error())
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal(val, &data); err != nil {
		r.log.Error(err.Error())

		return nil, err
	}

	return data, nil
}

func (r *redisCache) Key(network string, address string) string {
	return fmt.Sprintf("parser:%s:%s", network, address)
}

func NewRedisCache(redis *redis.Client, log *zap.Logger) *redisCache {
	return &redisCache{
		redis: redis,
		log:   log,
	}
}
