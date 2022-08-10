package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/kattana-io/tron-blocks-parser/internal/pair"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"go.uber.org/zap"
	"time"
)

type JMPairsCache struct {
	log   *zap.Logger
	redis Cache
	ttl   time.Duration
	api   *tronApi.Api
}

const KeyDiv = "jmcache"

func (c *JMPairsCache) GetPair(network string, address *tronApi.Address) (*pair.Pair, bool) {
	key := c.redis.Key(network, KeyDiv, address)
	ctx := context.Background()
	value, err := c.redis.Value(ctx, key)

	if err != nil || value == nil {
		pairEntity, ok := pair.NewPair(address, c.api, c.log)
		if ok {
			if err := c.redis.Store(ctx, key, &pairEntity, c.ttl); err != nil {
				c.log.Error(err.Error())
			}
			return &pairEntity, true
		} else {
			return &pair.Pair{}, false
		}
	} else {
		return value, true
	}
}

func CreateJMPairsCache(redis *redis.Client, api *tronApi.Api, log *zap.Logger) *JMPairsCache {
	return &JMPairsCache{
		api:   api,
		log:   log,
		redis: NewRedisCache(redis, log),
		ttl:   24 * time.Hour * 30, // 30 days
	}
}
