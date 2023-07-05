package pair

import (
	"github.com/kattana-io/tron-blocks-parser/internal/integrations"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/kattana-io/tron-objects-api/pkg/justmoney"
	"go.uber.org/zap"
)

type Pair struct {
	TokenA Token `json:"tokenA"`
	TokenB Token `json:"tokenB"`
}

const MaxRetry = 5

func NewPair(address *tronApi.Address, api *tronApi.API, tokenList *integrations.TokenListsProvider, log *zap.Logger) (Pair, bool) {
	pair := justmoney.New(api, *address)
	token0Addr, err := pair.Token0()
	if err != nil {
		log.Error("GetUniv2PairTokens: token0 " + err.Error())
		return Pair{}, false
	}
	token1Addr, err := pair.Token1()
	if err != nil {
		log.Error("GetUniv2PairTokens: token1 " + err.Error())
		return Pair{}, false
	}

	decimals, ok1 := tryGetTokenDecimals(token0Addr, api, tokenList, 0)
	decimals2, ok2 := tryGetTokenDecimals(token1Addr, api, tokenList, 0)

	if !ok1 || !ok2 {
		return Pair{}, false
	}

	return Pair{
		TokenA: Token{
			token0Addr.ToBase58(),
			int64(decimals),
		},
		TokenB: Token{
			token1Addr.ToBase58(),
			int64(decimals2),
		},
	}, true
}

func tryGetTokenDecimals(addr *tronApi.Address, api *tronApi.API, tokenList *integrations.TokenListsProvider, try int64) (int32, bool) {
	// for first time try to get from cache
	if try == 0 {
		dec, ok := tokenList.GetDecimals(addr)
		if ok {
			return dec, ok
		}
	}

	if try > MaxRetry {
		return 0, false
	}

	decimals, err := api.GetTokenDecimals(addr.ToHex())
	if err != nil {
		try += 1
		return tryGetTokenDecimals(addr, api, tokenList, try)
	} else {
		return decimals, true
	}
}
