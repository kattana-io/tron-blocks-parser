package parser

import (
	"github.com/ethereum/go-ethereum/common"
	commonModels "github.com/kattana-io/models/pkg/storage"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"math/big"
	"time"
)

func wrapETHAddress(input common.Address) *tronApi.Address {
	return tronApi.FromHex(input.String())
}

const swftswapProtocol = "swftswap"

func (p *Parser) onSwftSwap(log tronApi.Log, tx string, _ *tronApi.Address, timestamp int64) {
	event, err := p.abiHolder.SwftSwapAbi.EventByID(common.HexToHash(log.Topics[0]))

	if err != nil {
		p.log.Debug("Unpack error", zap.Error(err))
		return
	}
	data := make(map[string]any)
	if event != nil {
		// Unpack log into map
		err2 := event.Inputs.UnpackIntoMap(data, common.FromHex(log.Data))
		if err2 != nil {
			p.log.Debug("Unpack error", zap.Error(err2))
			return
		}
		fromAmount := data["fromAmount"].(*big.Int)
		fromToken := data["fromToken"].(common.Address)
		dstToken := common.HexToAddress(data["destination"].(string))
		minReturnAmount := data["minReturnAmount"].(*big.Int)
		sender := data["sender"].(common.Address)

		addrTokenA := wrapETHAddress(fromToken)
		addrTokenB := wrapETHAddress(dstToken)

		decimalsA, ok1 := p.GetTokenDecimals(addrTokenA)
		decimalsB, ok2 := p.GetTokenDecimals(addrTokenB)

		if !ok1 || !ok2 {
			p.log.Error("Could not get token decimals")
		}

		var priceA decimal.Decimal
		var priceB decimal.Decimal

		naturalA := decimal.NewFromBigInt(fromAmount, -decimalsA)
		naturalB := decimal.NewFromBigInt(minReturnAmount, -decimalsB)

		if !naturalA.IsZero() {
			priceA = naturalB.Div(naturalA)
		}

		if !naturalB.IsZero() {
			priceB = naturalA.Div(naturalB)
		}

		priceAUSD, priceBUSD := p.fiatConverter.ConvertAB(addrTokenA.ToBase58(), addrTokenB.ToBase58(), priceA)
		ValueUSD := calculateValueUSDSwftswap(naturalA, naturalB, priceAUSD, priceBUSD)

		dSwap := commonModels.DirectSwap{
			Tx:          tx,
			Date:        time.Unix(timestamp, 0),
			Chain:       Chain,
			BlockNumber: p.state.Block.Number.Uint64(),
			Protocol:    swftswapProtocol,
			SrcToken:    addrTokenA.ToBase58(),
			DstToken:    addrTokenB.ToBase58(),
			Amount0:     fromAmount,
			Amount1:     minReturnAmount,
			PriceA:      priceA,
			PriceAUSD:   priceAUSD,
			PriceB:      priceB,
			PriceBUSD:   priceBUSD,
			Wallet:      wrapETHAddress(sender).ToBase58(),
			Order:       0,
			ValueUSD:    ValueUSD,
		}
		p.state.AddDirectSwap(&dSwap)
	}
}

func calculateValueUSDSwftswap(amount0, amount1, ausd, busd decimal.Decimal) decimal.Decimal {
	if !ausd.IsZero() {
		return amount0.Mul(ausd)
	}
	if !busd.IsZero() {
		return amount1.Mul(busd)
	}
	return decimal.NewFromInt(0)
}
