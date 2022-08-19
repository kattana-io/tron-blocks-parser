package parser

import (
	"context"
	"github.com/kattana-io/tron-blocks-parser/internal/intermediate"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/kattana-io/tron-objects-api/pkg/trc20"
	"time"
)

const trxTokenAddress = "TRX"
const trxDecimals = 6

// GetCachePairToken - Try to pull from cache or populate cache
func (p *Parser) GetCachePairToken(address *tronApi.Address) (string, int32, bool) {
	pair, err := p.pairsCache.Value(context.Background(), address.ToBase58())
	if err != nil {
		pInstance := intermediate.Pair{Address: address.ToBase58()}
		tokenAddress, err := p.api.GetPairToken(address.ToHex())
		if err != nil {
			p.log.Error("GetCachePairToken: " + err.Error())
			return "", 0, false
		}
		// Check list
		decimals, ok := p.tokenLists.GetDecimals(tokenAddress)
		if ok {
			pInstance.SetToken(tokenAddress, decimals)
		} else {
			// Call API
			token := trc20.New(p.api, address)
			dec, err1 := token.GetDecimals()
			if err1 != nil {
				p.log.Error("GetCachePairToken: GetTokenDecimals: " + err.Error())
				return "", 0, false
			}
			pInstance.SetToken(tokenAddress, dec)
		}
		err2 := p.pairsCache.Store(context.Background(), address.ToBase58(), pInstance, time.Hour*2)
		if err2 != nil {
			p.log.Error("Could not put into cache: " + err.Error())
		}
		return pInstance.Token.Address, pInstance.Token.Decimals, true
	}
	return pair.Token.Address, pair.Token.Decimals, true
}

// GetPairTokens - Get tokens of pair
func (p *Parser) GetPairTokens(address *tronApi.Address) (string, int32, string, int32, bool) {
	adr, decimals, ok := p.GetCachePairToken(address)
	if ok {
		return adr, decimals, trxTokenAddress, trxDecimals, true
	}

	// Cache miss
	hexTokenAddress, err := p.api.GetPairToken(address.ToHex())
	tokenAddr := tronApi.FromHex(hexTokenAddress)
	cachedDecimals, ok := p.tokenLists.GetDecimals(tokenAddr.ToBase58())
	if ok {
		return tokenAddr.ToBase58(), cachedDecimals, trxTokenAddress, trxDecimals, true
	}

	token := trc20.New(p.api, address)
	decimals, err = token.GetDecimals()
	if err != nil {
		p.log.Error("GetPairTokens: " + err.Error())
		return "", 0, trxTokenAddress, trxDecimals, false
	}
	return tokenAddr.ToBase58(), decimals, trxTokenAddress, trxDecimals, true
}

func (p *Parser) GetUniv2PairTokens(address *tronApi.Address) (string, int32, string, int32, bool) {
	pair, ok := p.jmcache.GetPair(Chain, address)
	if ok {
		return pair.TokenA.Address, int32(pair.TokenA.Decimals), pair.TokenB.Address, int32(pair.TokenB.Decimals), true
	}

	return "", 0, "", 0, false
}
