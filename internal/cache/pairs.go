package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/intermediate"
	"go.uber.org/zap"
	"time"
)

type PairsCache struct {
	redis *redis.Client
	log   *zap.Logger
}

func (p *PairsCache) Key(address string) string {
	return fmt.Sprintf("parser:TRON:pair:%s", address)
}

func (p PairsCache) Store(ctx context.Context, address string, data intermediate.Pair, ttl time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		p.log.Error("Store: " + err.Error())
		return err
	}

	if err := p.redis.Set(ctx, p.Key(address), b, ttl).Err(); err != nil {
		p.log.Error("Store: " + err.Error())
		return err
	}

	return nil
}

func (p PairsCache) Value(ctx context.Context, address string) (*intermediate.Pair, error) {
	val, err := p.redis.Get(ctx, p.Key(address)).Bytes()
	if err == redis.Nil {
		return nil, err
	}
	if err != nil {
		p.log.Error("Value: " + err.Error())
		return nil, err
	}

	var data *intermediate.Pair
	if err := json.Unmarshal(val, &data); err != nil {
		p.log.Error("Value: " + err.Error())

		return nil, err
	}

	return data, nil
}

func (p *PairsCache) Warmup(pairs []intermediate.Pair) {
	for _, pair := range pairs {
		p.Store(context.Background(), pair.Address, pair, time.Hour*2)
	}
	p.log.Info(fmt.Sprintf("Added %d pairs", len(pairs)))
}

func NewPairsCache(redis *redis.Client, log *zap.Logger) *PairsCache {
	return &PairsCache{
		redis: redis,
		log:   log,
	}
}
