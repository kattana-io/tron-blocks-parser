package pair

import (
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	justmoney "github.com/kattana-io/tron-objects-api/pkg/justomoney"
	"go.uber.org/zap"
)

type Pair struct {
	TokenA Token `json:"tokenA"`
	TokenB Token `json:"tokenB"`
}

func NewPair(address *tronApi.Address, api *tronApi.Api, log *zap.Logger) (Pair, bool) {
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

	decimals := getTokenDecimals(token0Addr.ToBase58(), api, log)
	decimals2 := getTokenDecimals(token1Addr.ToBase58(), api, log)

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

func getTokenDecimals(addr string, api *tronApi.Api, log *zap.Logger) int32 {
	dec, err := api.GetTokenDecimals(addr)
	if err != nil {
		log.Error("getTokenDecimals: " + err.Error())
		return 0
	}
	return dec
}
