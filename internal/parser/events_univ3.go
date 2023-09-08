package parser

import (
	"github.com/ethereum/go-ethereum/common"
	commonModels "github.com/kattana-io/models/pkg/storage"
	abstractPair "github.com/kattana-io/tron-blocks-parser/internal/pair"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"math/big"
	"time"
)

func (p *Parser) onUniV3Swap(log tronApi.Log, tx string, owner *tronApi.Address, timestamp int64) {
	event, err := p.abiHolder.SwftSwapAbi.EventByID(common.HexToHash(log.Topics[0]))
	if err != nil {
		p.log.Debug("Unpack error", zap.Error(err))
		return
	}
	data := make(map[string]any)
	if event != nil {
		err2 := event.Inputs.UnpackIntoMap(data, common.FromHex(log.Data))
		if err2 != nil {
			p.log.Debug("Unpack error", zap.Error(err2))
			return
		}
		pair := tronApi.FromHex(log.Address)
		Amount0 := data["amount0"].(*big.Int)
		Amount1 := data["amount1"].(*big.Int)

		if Amount0.String() == "1" || Amount1.String() == "1" {
			p.log.Warn("Bad amounts, skipping: Token0Amount: " + Amount0.String() + " Token1Amount: " + Amount1.String())
			return
		}

		/**
		 * The delta of the token1 balance of the pool, if positive means amount1 added
		 * which means tokenA was bought
		 */
		BuyTokenA := decimal.NewFromBigInt(Amount1, 0).IsPositive()

		AmIn := decimal.NewFromBigInt(Amount0, 0).Abs()
		AmOut := decimal.NewFromBigInt(Amount1, 0).Abs()

		if AmIn.IsZero() || AmOut.IsZero() {
			return
		}

		tokenA, tokenB, ok := p.GetPairTokens(pair, abstractPair.UniV3)
		if !ok {
			p.log.Error("Could not dissolve univ2 pair: " + tx)
			return
		}

		naturalA := decimal.NewFromBigInt(Amount0, -tokenA.Decimals).Abs()
		naturalB := decimal.NewFromBigInt(Amount1, -tokenB.Decimals).Abs()

		PriceA := naturalB.Div(naturalA)
		PriceB := naturalA.Div(naturalB)

		PriceAUSD, PriceBUSD := p.fiatConverter.ConvertAB(tokenA.Address, tokenB.Address, PriceA)
		ValueUSD := p.calculateValueInUSD(Amount0, Amount1, pair, abstractPair.UniV3)

		trade := commonModels.PairSwap{
			Tx:          tx,
			Date:        time.Unix(timestamp, 0),
			Chain:       Chain,
			BlockNumber: p.state.Block.Number.Uint64(),
			Pair:        pair.ToBase58(),
			Amount0:     AmIn.BigInt(),
			Amount1:     AmOut.BigInt(),
			Buy:         BuyTokenA,
			PriceA:      PriceA,
			PriceAUSD:   PriceAUSD,
			PriceB:      PriceB,
			PriceBUSD:   PriceBUSD,
			ValueUSD:    ValueUSD,
			Wallet:      owner.ToBase58(),
		}
		p.state.AddTrade(&trade)
	} else {
		p.log.Debug("Could not unpack event, event is nil", zap.String("tx", tx))
		return
	}
}
