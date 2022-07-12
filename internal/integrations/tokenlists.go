package integrations

import (
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

/**
 * Method to parse token list and convert it into map
 * Why? to avoid extra calls to blockchains
 */

type Token struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	ChainId  int    `json:"chainId"`
	Decimals int    `json:"decimals"`
	Name     string `json:"name"`
	LogoURI  string `json:"logoURI"`
}

type TokenList struct {
	Name    string  `json:"name"`
	Tokens  []Token `json:"tokens"`
	LogoURI string  `json:"logoURI"`
	Version struct {
		Patch int `json:"patch"`
		Major int `json:"major"`
		Minor int `json:"minor"`
	} `json:"version"`
	Timestamp int64 `json:"timestamp"`
}

const url = "https://list.justswap.link/justswap.json"

func fetch() (data *TokenList, err error) {
	res, err := http.Get(url)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(data)
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
	resp, err := fetch()
	ok := err == nil

	return &TokenListsProvider{
		log:      log,
		ok:       ok,
		decimals: createDecimalsList(resp),
	}
}

func createDecimalsList(resp *TokenList) *sync.Map {
	smp := sync.Map{}
	for _, token := range resp.Tokens {
		smp.Store(token.Address, token.Decimals)
	}
	return &smp
}

// GetDecimals - Fetch decimals from sync map
func (t *TokenListsProvider) GetDecimals(address string) (int32, bool) {
	val, ok := t.decimals.Load(address)
	if ok {
		return int32(val.(int)), true
	}
	return 0, false
}
