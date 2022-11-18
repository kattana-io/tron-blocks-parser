package parser

import (
	"context"
	"fmt"
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
		hexTokenAddress, err := p.api.GetPairToken(address.ToHex())
		if err != nil {
			p.log.Error("GetCachePairToken: " + err.Error())
			return "", 0, false
		}
		if hexTokenAddress == "" {
			p.log.Error("Couldn't get token address for pair: " + address.ToBase58())
			return "", 0, false
		}
		tokenAddr := tronApi.FromHex(hexTokenAddress)
		// Check list
		decimals, ok := p.tokenLists.GetDecimals(tokenAddr)
		if ok {
			pInstance.SetToken(tokenAddr.ToBase58(), decimals)
		} else {
			// Call API
			token := trc20.New(p.api, tokenAddr)
			dec, ok := token.TryToGetDecimals(0)
			if !ok {
				p.log.Error("TryToGetDecimals: tried 5 times w/o result")
				return "", 0, false
			}
			pInstance.SetToken(tokenAddr.ToBase58(), dec)
		}
		err2 := p.pairsCache.Store(context.Background(), address.ToBase58(), pInstance, time.Hour*2)
		if err2 != nil {
			p.log.Error("Could not put into cache: " + err2.Error())
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
	if err != nil {
		p.log.Error(fmt.Sprintf("GetPairToken: %s", err.Error()))
		return "", 0, trxTokenAddress, trxDecimals, false
	}
	if hexTokenAddress == "" {
		p.log.Error("Couldn't get token address for pair: " + address.ToBase58())
		return "", 0, trxTokenAddress, trxDecimals, false
	}
	tokenAddr := tronApi.FromHex(hexTokenAddress)
	cachedDecimals, ok := p.tokenLists.GetDecimals(tokenAddr)
	if ok {
		return tokenAddr.ToBase58(), cachedDecimals, trxTokenAddress, trxDecimals, true
	}

	token := trc20.New(p.api, tokenAddr)
	decimals, ok = token.TryToGetDecimals(0)
	if !ok {
		p.log.Error("TryToGetDecimals: tried 5 times w/o result")
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
