package converters

import (
	"context"
	"github.com/shopspring/decimal"
)

/**
 * Contains logic to fetch last prices
 */

const LiveCacheKey = "parser:prices:TRON:live"

// readLastPrices - read last prices from hash of live prices
func (f *FiatConverter) readLastPrices() {
	f.ratesMutex.Lock()
	defer f.ratesMutex.Unlock()
	result := f.redis.HGetAll(context.Background(), LiveCacheKey).Val()

	for key, value := range result {
		price, err := decimal.NewFromString(value)
		if err == nil {
			f.Prices[key] = price
		}
	}
}

func (f *FiatConverter) writeLastPrices() {
	f.ratesMutex.Lock()
	defer f.ratesMutex.Unlock()

	result := make(map[string]string, 0)
	for key, value := range f.Prices {
		result[key] = value.String()
	}
	f.redis.HSet(context.Background(), LiveCacheKey, result)
}
