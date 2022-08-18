package integrations

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-objects-api/pkg/api"
	"go.uber.org/zap"
	"io/ioutil"
	"sync"
)

/**
 * Method to parse token list and convert it into map
 * Why? to avoid extra calls to blockchains
 */

type Token struct {
	Decimals int    `json:"decimals"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
}

func fetch() (map[string]Token, error) {
	raw, err := ioutil.ReadFile("tokens.json")
	if err != nil {
		return nil, err
	}
	data := make(map[string]Token, 0)
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return data, err
}

type TokenListsProvider struct {
	log      *zap.Logger
	ok       bool
	decimals *sync.Map
}

func NewTokensListProvider(log *zap.Logger) *TokenListsProvider {
	list, err := fetch()
	ok := err == nil
	if ok {
		log.Info(fmt.Sprintf("Loaded %d tokens", len(list)))
	}

	return &TokenListsProvider{
		log:      log,
		ok:       ok,
		decimals: createDecimalsList(list),
	}
}

// ensure that we have check address
func normalizeAddress(address string) string {
	return api.FromBase58(address).ToBase58()
}

func createDecimalsList(resp map[string]Token) *sync.Map {
	smp := sync.Map{}
	if resp != nil {
		for key, token := range resp {
			smp.Store(normalizeAddress(key), token.Decimals)
		}
	}
	return &smp
}

// GetDecimals - Fetch decimals from sync map
func (t *TokenListsProvider) GetDecimals(address string) (int32, bool) {
	val, ok := t.decimals.Load(normalizeAddress(address))
	if ok {
		return int32(val.(int)), true
	}
	return 0, false
}
