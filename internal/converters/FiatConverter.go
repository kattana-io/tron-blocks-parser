package converters

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"io/ioutil"
	"math/big"
	"sync"
	"time"
)

type FiatConverter struct {
	redis            *redis.Client
	Prices           map[string]decimal.Decimal `json:"prices"`
	supported        map[string]bool
	pairs            map[string]bool
	flips            map[string]bool
	stableCoinsList  map[string]bool
	rawQuotes        []models.QuotePair
	stableCoinsMutex sync.Mutex
	ratesMutex       sync.RWMutex
	listMutex        sync.RWMutex
	block            *models.Block
	log              *zap.Logger
}

const StableCoin = 2

// ReadFile and update internal state
func (f *FiatConverter) ReadFile() {
	raw, err := ioutil.ReadFile("quotes.json")
	if err != nil {
		return
	}
	cFile := models.ConfigFile{}
	err = json.Unmarshal(raw, &cFile)
	if err != nil {
		return
	}
	f.rawQuotes = cFile.Quotes
}

func CreateConverter(client *redis.Client, log *zap.Logger, block *models.Block) *FiatConverter {
	converter := &FiatConverter{
		log:             log,
		redis:           client,
		block:           block,
		Prices:          make(map[string]decimal.Decimal, 0),
		flips:           make(map[string]bool, 0),
		supported:       make(map[string]bool, 0),
		pairs:           make(map[string]bool, 0),
		stableCoinsList: make(map[string]bool, 0),
		rawQuotes:       make([]models.QuotePair, 0),
	}

	converter.ReadFile()

	for _, quote := range converter.rawQuotes {
		if quote.Kind == StableCoin {
			converter.Prices[quote.Token] = decimal.NewFromInt(1)
			converter.stableCoinsList[quote.Token] = true
		}
		converter.supported[quote.Token] = true
	}

	converter.readPreviousBlockPricesFromCache()
	return converter
}

// check before & after condition
func (f *FiatConverter) skipQuote(quote models.QuotePair) bool {
	if quote.After != nil {
		return f.block.Number.Cmp(quote.After) == -1
	}
	if quote.Before != nil {
		return f.block.Number.Cmp(quote.Before) == 1
	}
	return false
}

func (f *FiatConverter) Update(pair string, tokenA string, tokenB string, price decimal.Decimal) {
	if f.updateable(pair) {
		if f.ShouldFlip(pair) {
			f.updatePrice(tokenB, price, true)
		} else {
			f.updatePrice(tokenA, price, false)
		}
	}
}

func (f *FiatConverter) Convertable(address string) bool {
	f.listMutex.RLock()
	defer f.listMutex.RUnlock()
	if _, ok := f.supported[address]; ok {
		return true
	}
	return false
}

func (f *FiatConverter) ShouldFlip(address string) bool {
	f.listMutex.RLock()
	defer f.listMutex.RUnlock()
	if val, ok := f.flips[address]; ok {
		return val
	}
	return false
}

func (f *FiatConverter) GetPriceOfToken(token string) decimal.Decimal {
	if f.isTokenStable(token) {
		return decimal.NewFromInt(1)
	}
	if f.Convertable(token) {
		return f.getRate(token)
	}
	return decimal.NewFromInt(0)
}

func (f *FiatConverter) Convert(tokenA string, tokenB string, price decimal.Decimal) decimal.Decimal {
	if f.isTokenStable(tokenA) {
		if !price.IsZero() {
			return decimal.NewFromInt(1).Div(price)
		}
	}
	if f.Convertable(tokenB) {
		return price.Mul(f.getRate(tokenB))
	}
	return decimal.NewFromInt(0)
}

// ConvertAB - return both prices
func (f *FiatConverter) ConvertAB(tokenA string, tokenB string, price decimal.Decimal) (priceAUSD decimal.Decimal, priceBUSD decimal.Decimal) {
	if !price.IsZero() {
		if f.isTokenStable(tokenA) {
			return decimal.NewFromInt(1), decimal.NewFromInt(1).Div(price)
		}
		if f.isTokenStable(tokenB) {
			return price, decimal.NewFromInt(1)
		}
	}
	if f.Convertable(tokenB) {
		rateB := f.getRate(tokenB)
		return price.Mul(rateB), rateB
	}
	if f.Convertable(tokenA) {
		rateA := f.getRate(tokenA)
		if !price.IsZero() {
			return rateA, rateA.Div(price)
		}
		return rateA, decimal.NewFromInt(0)
	}
	return
}

func (f *FiatConverter) Commit() {
	b, _ := json.Marshal(f)

	key := cacheKey(f.block.Network, f.block.Number.String())

	if err := f.redis.Set(context.Background(), key, b, time.Second*30).Err(); err != nil {
		f.log.Error(err.Error())
	}
}

func (f *FiatConverter) updateable(pair string) bool {
	f.listMutex.RLock()
	defer f.listMutex.RUnlock()
	if _, ok := f.pairs[pair]; ok {
		return true
	}
	return false
}

func (f *FiatConverter) updatePrice(token string, price decimal.Decimal, flipped bool) {
	f.ratesMutex.Lock()
	defer f.ratesMutex.Unlock()

	if flipped {
		f.Prices[token] = decimal.NewFromInt(1).Div(price)
	} else {
		f.Prices[token] = price
	}
}

func (f *FiatConverter) getRate(token string) decimal.Decimal {
	defer f.ratesMutex.RUnlock()
	f.ratesMutex.RLock()

	if val, ok := f.Prices[token]; ok {
		return val
	}
	return decimal.NewFromInt(0)
}

func (f *FiatConverter) isTokenStable(token string) bool {
	defer f.stableCoinsMutex.Unlock()
	f.stableCoinsMutex.Lock()

	if f.stableCoinsList[token] {
		return true
	}
	return false
}

/**
 * Cache interaction methods
 */

func cacheKey(network string, number string) string {
	return fmt.Sprintf("parser:prices:%s:%s", network, number)
}

func (f *FiatConverter) readPreviousBlockPricesFromCache() bool {
	blockNumber := big.NewInt(0).Sub(f.block.Number, big.NewInt(1)) // previous block
	key := cacheKey(f.block.Network, blockNumber.String())

	val, err := f.redis.Get(context.Background(), key).Bytes()
	if err == redis.Nil || err != nil {
		return false
	}

	cached := &FiatConverter{}
	if err := json.Unmarshal(val, &cached); err != nil {
		f.log.Error(err.Error())
		return false
	}

	f.Prices = cached.Prices
	return true
}
