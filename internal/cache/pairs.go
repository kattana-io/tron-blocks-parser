package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"go.uber.org/zap"
	"time"
)

type PairCache interface {
	Set(context.Context, string, *models.Pair) error
	Get(context.Context, string) (*models.Pair, error)
}

// pair cache ttl
const ttl = time.Hour * 96

type RedisPairsCache struct {
	redis *redis.Client
}

func (p *RedisPairsCache) Set(ctx context.Context, address string, pair *models.Pair) error {
	b, err := json.Marshal(pair)
	if err != nil {
		zap.L().Error("Set, json", zap.Error(err))
		return err
	}

	if err = p.redis.Set(ctx, p.Key(address), b, ttl).Err(); err != nil {
		zap.L().Error("Set, write to redis", zap.Error(err))
		return err
	}

	return nil
}

func (p *RedisPairsCache) Get(ctx context.Context, address string) (*models.Pair, error) {
	val, err := p.redis.Get(ctx, p.Key(address)).Bytes()
	if err == redis.Nil {
		return nil, err
	}
	if err != nil {
		zap.L().Error("get: ", zap.Error(err))
		return nil, err
	}

	var data models.Pair
	if err := json.Unmarshal(val, &data); err != nil {
		zap.L().Error("get: ", zap.Error(err))
		return nil, err
	}

	return &data, nil
}

func (p *RedisPairsCache) Key(address string) string {
	return fmt.Sprintf("parser:TRON:pair:v2:%s", address)
}

func NewPairsCache(redis *redis.Client) PairCache {
	return &RedisPairsCache{
		redis: redis,
	}
}
