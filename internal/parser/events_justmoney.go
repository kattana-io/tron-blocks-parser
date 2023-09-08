package parser

import (
	abstractPair "github.com/kattana-io/tron-blocks-parser/internal/pair"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	commonModels "github.com/kattana-io/models/pkg/storage"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

//nolint:lll
const JMFactoryABI = `[{"inputs":[{"internalType":"address","name":"_owner","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"_moderator","type":"address"}],"name":"addModerator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"allPairs","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"allPairsLength","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"}],"name":"createPair","outputs":[{"internalType":"address","name":"pair","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"feeTo","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"address","name":"","type":"address"}],"name":"getPair","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"pair","type":"address"}],"name":"getPairSymbols","outputs":[{"internalType":"string","name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getRouterAddress","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"getTokensByPair","outputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_moderator","type":"address"}],"name":"removeModerator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_feeTo","type":"address"}],"name":"setFeeTo","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_newOwner","type":"address"}],"name":"setNewOwner","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_router","type":"address"}],"name":"setRouterAddress","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

func (p *Parser) onJmPairCreated(log tronApi.Log, timestamp int64) {
	factory := tronApi.FromHex(log.Address)

	factoryAbi, err := abi.JSON(strings.NewReader(JMFactoryABI))
	if err != nil {
		p.log.Warn("Could not parse factory abi: ", zap.Error(err))
		return
	}

	event, err := factoryAbi.EventByID(common.HexToHash(log.Topics[0]))
	if event != nil {
		if err != nil {
			p.log.Debug("Unpack error", zap.Error(err))
			return
		}
		data := make(map[string]any)
		err2 := event.Inputs.UnpackIntoMap(data, common.FromHex(log.Data))
		if err2 != nil {
			p.log.Debug("Unpack error", zap.Error(err2))
			return
		}
		pairAddress := data["pair"].(common.Address)
		pair := tronApi.FromHex(pairAddress.Hex())
		nodeURL := os.Getenv("FULL_NODE_URL")
		p.state.RegisterNewPair(factory.ToBase58(), pair.ToBase58(), "justmoney", Chain, nodeURL, time.Unix(timestamp, 0))
	}
}

// GetUniV2Buy Returns Buy true/false
func GetUniV2Buy(amount0In, amount0Out, amount1In, amount1Out *big.Int) bool {
	zero := big.NewInt(0)
	// tokenB amount is 0, so we are selling tokenA
	if amount1In.Cmp(zero) == 0 {
		return false
	}
	// tokenA is 0, so we are buying it
	if amount0In.Cmp(zero) == 0 {
		return true
	}
	if amount1In.Cmp(zero) == 1 && amount0Out.Cmp(zero) == 1 {
		return true // Buy
	}
	if amount0In.Cmp(zero) == 1 && amount1Out.Cmp(zero) == 1 {
		return false // Sell
	}
	return false
}

func (p *Parser) onJmSyncEvent(log tronApi.Log, tx string, owner *tronApi.Address, timestamp int64) {
	event, err := p.abiHolder.JMPairAbi.EventByID(common.HexToHash(log.Topics[0]))

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
		pair := tronApi.FromHex(log.Address)

		reserves0 := data["reserve0"].(*big.Int)
		reserves1 := data["reserve1"].(*big.Int)

		tokenA, tokenB, ok := p.GetPairTokens(pair, abstractPair.UniV2)
		if !ok {
			p.log.Error("Could not dissolve univ2 pair: " + tx)
			return
		}

		// @todo verify price formula
		priceA := decimal.Decimal{}
		res0 := decimal.NewFromBigInt(reserves0, -tokenA.Decimals)
		if !res0.IsZero() {
			priceA = decimal.NewFromBigInt(reserves1, -tokenB.Decimals).Div(res0)
		}

		priceB := decimal.Decimal{}
		res1 := decimal.NewFromBigInt(reserves1, -tokenB.Decimals)
		if !res1.IsZero() {
			priceB = decimal.NewFromBigInt(reserves0, -tokenA.Decimals).Div(res1)
		}

		priceAUSD, priceBUSD := p.fiatConverter.ConvertAB(tokenA.Address, tokenB.Address, priceA)
		reservesUSD := p.calculateReservesInUSD(reserves0, reserves1, pair, abstractPair.UniV2)

		sync := commonModels.LiquidityEvent{
			BlockNumber: p.state.Block.Number.Uint64(),
			Date:        time.Unix(timestamp, 0),
			Tx:          tx,
			Pair:        pair.ToBase58(),
			Chain:       Chain,
			Klass:       "sync",
			Wallet:      owner.ToBase58(),
			Order:       0,
			Reserve0:    reserves0.String(),
			Reserve1:    reserves1.String(),
			PriceA:      priceA,
			PriceAUSD:   priceAUSD,
			PriceB:      priceB,
			PriceBUSD:   priceBUSD,
			ReserveUSD:  reservesUSD,
		}

		p.state.AddLiquidity(&sync)
	}
}

func (p *Parser) onJmSwapEvent(log tronApi.Log, tx string, owner *tronApi.Address, timestamp int64) {
	event, err := p.abiHolder.JMPairAbi.EventByID(common.HexToHash(log.Topics[0]))
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
		pair := tronApi.FromHex(log.Address)
		Amount0In := data["amount0In"].(*big.Int)
		Amount1Out := data["amount1Out"].(*big.Int)
		Amount1In := data["amount1In"].(*big.Int)
		Amount0Out := data["amount0Out"].(*big.Int)

		if Amount0In.String() == "1" || Amount1In.String() == "1" {
			p.log.Warn("Bad amounts, skipping: Token0Amount: " + Amount0In.String() + " Token1Amount: " + Amount1In.String())
			return
		}

		Token0Amount := big.NewInt(0).Abs(big.NewInt(0).Sub(Amount0In, Amount0Out))
		Token1Amount := big.NewInt(0).Abs(big.NewInt(0).Sub(Amount1Out, Amount1In))

		Buy := GetUniV2Buy(Amount0In, Amount0Out, Amount1In, Amount1Out)

		AmIn := decimal.NewFromBigInt(Token0Amount, 0)
		AmOut := decimal.NewFromBigInt(Token1Amount, 0)

		if AmIn.IsZero() || AmOut.IsZero() {
			p.log.Warn("Amounts of swap are zeroes, skipping", zap.Error(err))
			return
		}

		tokenA, tokenB, ok := p.GetPairTokens(pair, abstractPair.UniV2)
		if !ok {
			p.log.Error("Could not dissolve univ2 pair: " + tx)
			return
		}

		naturalA := decimal.NewFromBigInt(Token0Amount, -tokenA.Decimals).Abs()
		naturalB := decimal.NewFromBigInt(Token1Amount, -tokenB.Decimals).Abs()

		PriceA := naturalB.Div(naturalA)
		PriceB := naturalA.Div(naturalB)

		PriceAUSD, PriceBUSD := p.fiatConverter.ConvertAB(tokenA.Address, tokenB.Address, PriceA)
		ValueUSD := p.calculateValueInUSD(Token0Amount, Token1Amount, pair, abstractPair.UniV2)

		trade := commonModels.PairSwap{
			Tx:          tx,
			Date:        time.Unix(timestamp, 0),
			Chain:       Chain,
			BlockNumber: p.state.Block.Number.Uint64(),
			Pair:        pair.ToBase58(),
			Amount0:     Token0Amount,
			Amount1:     Token1Amount,
			Buy:         Buy,
			PriceA:      PriceA,
			PriceAUSD:   PriceAUSD,
			PriceB:      PriceB,
			PriceBUSD:   PriceBUSD,
			ValueUSD:    ValueUSD,
			Bot:         false,
			Wallet:      owner.ToBase58(),
			Order:       0,
		}

		p.state.AddTrade(&trade)
		return
	} else {
		p.log.Debug("Could not unpack event, event is nil: " + tx)
		return
	}
}
