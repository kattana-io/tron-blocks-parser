package parser

import (
	"context"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	abstractPair "github.com/kattana-io/tron-blocks-parser/internal/pair"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	jmPair "github.com/kattana-io/tron-objects-api/pkg/justmoney"
	"go.uber.org/zap"
)

const (
	trxAddress   = "TNUC9Qb1rRpS5CbWLmNMxXBjyFoydXjWFR"
	trxDecimals  = 6
	brokenPair   = "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE"
	usdtAddress  = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	usdtDecimals = 6
)

func (p *Parser) GetPairTokens(pair *tronApi.Address, klass string) (tokenA, tokenB *models.Token, ok bool) {
	ctx := context.Background()
	// Step 1: Check if pair is present in cache
	instance, err := p.pairsCache.Get(ctx, pair.ToBase58())
	if err != nil {
		// Step 2: Create pair instance
		inst, ok := p.CreatePair(ctx, pair, klass)
		if ok {
			err = p.pairsCache.Set(ctx, pair.ToBase58(), inst)
			if err != nil {
				p.log.Error("GetPairTokens, set pair", zap.Error(err))
			}
			return &inst.Token0, &inst.Token1, true
		}
		// Step 3: If we failed to fetch than return nil
		return nil, nil, false
	}
	return &instance.Token0, &instance.Token1, true
}

func (p *Parser) createToken(address *tronApi.Address) models.Token {
	// Step 1: fetch from cached token list
	dec, ok := p.tokenLists.GetDecimals(address)
	if ok {
		p.log.Info("fetched decimals from list for token ", zap.String("address", address.ToBase58()))
		return models.Token{
			Address:  address.ToBase58(),
			Decimals: dec,
		}
	}
	// Step 2: do a static call for trc20 token
	dec, err := p.api.GetTokenDecimals(address.ToHex())
	if err != nil {
		p.log.Error("createToken: GetTokenDecimals", zap.Error(err))
	}
	return models.Token{
		Address:  address.ToBase58(),
		Decimals: dec,
	}
}

func (p *Parser) CreatePair(_ context.Context, addr *tronApi.Address, klass string) (*models.Pair, bool) {
	switch klass {
	case abstractPair.UniV2:
	case abstractPair.UniV3: // uniV3 same function names
		instance := jmPair.New(p.api, *addr)
		addr0, err := instance.Token0()
		if err != nil {
			p.log.Error("could not fetch token0",
				zap.Error(err),
				zap.String("pair", addr.ToBase58()))
			return nil, false
		}
		addr1, err := instance.Token1()
		if err != nil {
			p.log.Error("could not fetch token1",
				zap.Error(err),
				zap.String("pair", addr.ToBase58()))
			return nil, false
		}
		return &models.Pair{
			Address: addr.ToBase58(),
			Klass:   klass,
			Token0:  p.createToken(addr0),
			Token1:  p.createToken(addr1),
		}, true
	case abstractPair.Sunswap:
		pair := models.Pair{
			Address: addr.ToBase58(),
			Klass:   abstractPair.Sunswap,
			Token1: models.Token{
				Address:  trxAddress,
				Decimals: trxDecimals,
			},
		}
		// Spike for broken pair
		if addr.ToBase58() == brokenPair {
			pair.Token0 = models.Token{
				Address:  usdtAddress,
				Decimals: usdtDecimals,
			}
			return &pair, true
		}
		// Default flow
		addr0, ok := p.GetSunswapToken(addr)
		if !ok {
			p.log.Error("Sunswap: could not fetch tokenAddress",
				zap.String("pair", addr.ToBase58()))
			return nil, false
		}
		if addr0 == "" {
			return nil, false
		}
		tokenAddr := tronApi.FromHex(addr0)
		pair.Token0 = p.createToken(tokenAddr)

		return &pair, true
	default:
		p.log.Error("unknown pair type", zap.String("klass", klass))
		return nil, false
	}
	return nil, false
}

func (p *Parser) GetSunswapToken(addr *tronApi.Address) (string, bool) {
	data, err := p.api.TCCRequest(map[string]any{
		"contract_address":  addr.ToHex(),
		"owner_address":     "4128fb7be6c95a27217e0e0bff42ca50cd9461cc9f",
		"function_selector": "tokenAddress()",
		"parameter":         "",
		"call_value":        0,
	})

	if err != nil || len(data.ConstantResult) == 0 {
		return "", false
	}

	return tronApi.TrimZeroes(data.ConstantResult[0]), true
}
