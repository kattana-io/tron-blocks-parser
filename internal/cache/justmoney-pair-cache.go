package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kattana-io/tron-blocks-parser/internal/integrations"
	"github.com/kattana-io/tron-blocks-parser/internal/pair"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"go.uber.org/zap"
)

type JMPairsCache struct {
	log       *zap.Logger
	redis     Cache
	ttl       time.Duration
	api       *tronApi.API
	tokenList *integrations.TokenListsProvider
}

const KeyDiv = "jmcache"

func (c *JMPairsCache) GetPair(network string, address *tronApi.Address) (*pair.Pair, bool) {
	key := c.redis.Key(network, KeyDiv, address)
	ctx := context.Background()
	value, err := c.redis.Value(ctx, key)

	if err != nil || value == nil {
		pairEntity, ok := pair.NewPair(address, c.api, c.tokenList, c.log)
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

func CreateJMPairsCache(redis *redis.Client, api *tronApi.API, tokenList *integrations.TokenListsProvider, log *zap.Logger) *JMPairsCache {
	return &JMPairsCache{
		api:       api,
		log:       log,
		redis:     NewRedisCache(redis, log),
		tokenList: tokenList,
		ttl:       24 * time.Hour * 30, // 30 days
	}
}
