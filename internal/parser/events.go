package parser

/**
 * Package with event handling logic
 */

import (
	"github.com/kattana-io/tron-blocks-parser/internal/helper"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"sync"
	"time"
)

const Chain = "TRON"

/**
 * List of supported events
 */
const transferEvent = 0xddf252ad
const liquidityAdded = 0x06239653
const liquidityRemoved = 0x0fbf06c0
const tokenPurchaseEvent = 0xcd60aa75
const trxPurchaseEvent = 0xdad9ec5c
const snapshotEvent = 0xcc7244d3
const listingEvent = 0x9d42cb01
const jmListingEvent = 0x0d3648bd
const jmUniv2SwapEvent = 0xd78ad95f

func (p *Parser) processLog(log tronApi.Log, tx string, timestamp int64, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(log.Topics) < 1 {
		return
	}
	methodId := getMethodId(log.Topics[0])
	switch methodId {
	//case transferEvent:
	//	p.onTokenTransfer(log, tx, timestamp)
	case tokenPurchaseEvent:
		p.onTokenPurchase(log, tx, timestamp)
	case trxPurchaseEvent:
		p.onTrxPurchase(log, tx, timestamp)
	case snapshotEvent:
		p.onPairSnapshot(log, tx, timestamp)
	case listingEvent:
		p.onPairCreated(log, tx, timestamp)
	case jmListingEvent:
		p.onJmPairCreated(log, tx, timestamp)
	case jmUniv2SwapEvent:
		p.onJmSwapEvent(log, tx, timestamp)
	}
}

// topics - from, to, value
func (p *Parser) onTokenTransfer(log tronApi.Log, tx string, timestamp int64) {
	Contract := tronApi.FromHex(log.Address)
	from := log.Topics[0]
	to := log.Topics[1]
	value := log.Topics[2]

	tokenA, decimals0, tokenB, decimals1, ok := p.GetPairTokens(Contract)
	if !ok {
		p.log.Error("Could not dissolve pair: onTokenPurchase")
		return
	}

	// Normalize amounts
	trxAmountRaw, err1 := decimal.NewFromString(log.Topics[2])
	if err1 != nil {
		p.log.Error("Could not parse amounts" + err1.Error())
		return
	}
	tokenAmountRaw, err2 := decimal.NewFromString(log.Topics[3])
	if err2 != nil {
		p.log.Error("Could not parse amounts" + err2.Error())
		return
	}

	// Convert to natural amounts by dropping decimals
	trxAmount := trxAmountRaw.Div(decimal.New(1, decimals0))
	tokenAmount := tokenAmountRaw.Div(decimal.New(1, decimals1))
	// Calculate prices
	priceA := tokenAmount.Div(trxAmount)
	priceAUSD, priceBUSD := p.fiatConverter.ConvertAB(tokenA, tokenB, priceA)

	tEvent := models.TransferEvent{
		Chain:       Chain,
		Contract:    Contract.ToBase58(),
		BlockNumber: p.state.Block.Number.Uint64(),
		Date:        time.Unix(timestamp, 0),
		Order:       0,
		Tx:          tx,
		From:        from,
		To:          to,
		Amount:      value,
		ValueUSD:    calculateValueUSD(tokenAmount, trxAmount, priceAUSD, priceBUSD),
	}
	p.state.AddTransferEvent(&tEvent)
}

// topics - buyer,trx_sold,tokens_bought
func (p *Parser) onTokenPurchase(log tronApi.Log, tx string, timestamp int64) {
	pair := tronApi.FromHex(log.Address)
	buyer := tronApi.TrimZeroes(log.Topics[1])
	// Dissolve pair
	tokenA, decimals0, tokenB, decimals1, ok := p.GetPairTokens(pair)

	if !ok {
		p.log.Error("Could not dissolve pair: onTokenPurchase")
	}

	// Normalize amounts
	trxAmountRaw := helper.TronValueToDecimal(log.Topics[2])
	tokenAmountRaw := helper.TronValueToDecimal(log.Topics[3])

	// Convert to natural amounts by dropping decimals
	trxAmount := trxAmountRaw.Div(decimal.New(1, decimals1))
	tokenAmount := tokenAmountRaw.Div(decimal.New(1, decimals0))

	if tokenAmount.IsZero() || trxAmount.IsZero() {
		p.log.Warn("Skipping division by zero")
		return
	}
	// Calculate prices
	priceA := trxAmount.Div(tokenAmount)
	priceB := tokenAmount.Div(trxAmount)

	priceAUSD, priceBUSD := p.fiatConverter.ConvertAB(tokenA, tokenB, priceA)
	valueUSD := calculateValueUSD(tokenAmount, trxAmount, priceAUSD, priceBUSD)

	swap := models.PairSwap{
		Tx:          tx,
		Date:        time.Unix(timestamp, 0),
		Chain:       Chain,
		BlockNumber: p.state.Block.Number.Uint64(),
		Pair:        pair.ToBase58(),
		Amount0:     tokenAmountRaw.BigInt(),
		Amount1:     trxAmountRaw.BigInt(),
		Buy:         true,
		PriceA:      priceA,
		PriceAUSD:   priceAUSD,
		PriceB:      priceB,
		PriceBUSD:   priceBUSD,
		Bot:         false,
		Wallet:      tronApi.FromHex(buyer).ToBase58(),
		Order:       0,
		ValueUSD:    valueUSD,
	}
	p.state.AddTrade(&swap)
}

// topics - buyer, tokens_sold, trx_bought
func (p *Parser) onTrxPurchase(log tronApi.Log, tx string, timestamp int64) {
	pair := tronApi.FromHex(log.Address)
	buyer := tronApi.TrimZeroes(log.Topics[1])
	// Dissolve pair
	tokenA, decimals0, tokenB, decimals1, ok := p.GetPairTokens(pair)

	if !ok {
		p.log.Error("Could not dissolve tokens: onTrxPurchase")
		return
	}
	// Normalize amounts
	tokenAmountRaw := helper.TronValueToDecimal(log.Topics[2])
	trxAmountRaw := helper.TronValueToDecimal(log.Topics[3])

	// Convert to natural amounts by dropping decimals
	tokenAmount := tokenAmountRaw.Div(decimal.New(1, decimals0))
	trxAmount := trxAmountRaw.Div(decimal.New(1, decimals1))

	if tokenAmount.IsZero() || trxAmount.IsZero() {
		p.log.Warn("Skipping division by zero")
		return
	}

	// Calculate prices
	priceA := trxAmount.Div(tokenAmount)
	priceB := tokenAmount.Div(trxAmount)

	priceAUSD, priceBUSD := p.fiatConverter.ConvertAB(tokenA, tokenB, priceA)
	valueUSD := calculateValueUSD(tokenAmount, trxAmount, priceAUSD, priceBUSD)

	swap := models.PairSwap{
		Tx:          tx,
		Date:        time.Unix(timestamp, 0),
		Chain:       Chain,
		BlockNumber: p.state.Block.Number.Uint64(),
		Pair:        pair.ToBase58(),
		Amount0:     tokenAmountRaw.BigInt(),
		Amount1:     trxAmountRaw.BigInt(),
		Buy:         false,
		PriceA:      priceA,
		PriceAUSD:   priceAUSD,
		PriceB:      priceB,
		PriceBUSD:   priceBUSD,
		Bot:         false,
		Wallet:      tronApi.FromHex(buyer).ToBase58(),
		Order:       0,
		ValueUSD:    valueUSD,
	}
	p.state.AddTrade(&swap)
}

func calculateValueUSD(amount0 decimal.Decimal, amount1 decimal.Decimal, ausd decimal.Decimal, busd decimal.Decimal) decimal.Decimal {
	if !ausd.IsZero() {
		return amount0.Mul(ausd)
	}
	if !busd.IsZero() {
		return amount1.Mul(busd)
	}
	return decimal.NewFromInt(0)
}

// Snapshot event to sync liquidity
// topics - operator, trx_balance, token_balance
func (p *Parser) onPairSnapshot(log tronApi.Log, tx string, timestamp int64) {
	if len(log.Topics) != 4 {
		p.log.Error("onPairSnapshot: Invalid length of topics")
		return
	}
	pair := tronApi.FromHex(log.Address)
	operator := tronApi.TrimZeroes(log.Topics[1])

	// Dissolve pair
	tokenA, decimals0, tokenB, decimals1, ok := p.GetPairTokens(pair)

	if !ok {
		p.log.Error("Could not dissolve tokens: onTrxPurchase")
		return
	}
	// Normalize amounts
	trxAmountRaw := helper.TronValueToDecimal(log.Topics[2])
	tokenAmountRaw := helper.TronValueToDecimal(log.Topics[3])

	// Convert to natural amounts by dropping decimals
	tokenAmount := tokenAmountRaw.Div(decimal.New(1, decimals0))
	trxAmount := trxAmountRaw.Div(decimal.New(1, decimals1))

	if tokenAmount.IsZero() || trxAmount.IsZero() {
		p.log.Warn("Skipping division by zero")
		return
	}

	// Calculate prices
	priceA := trxAmount.Div(tokenAmount)

	priceAUSD, priceBUSD := p.fiatConverter.ConvertAB(tokenA, tokenB, priceA)
	p.fiatConverter.UpdateTokenUSDPrice(tokenA, priceAUSD)

	if p.isPairWhiteListed(pair) {
		p.fiatConverter.UpdateTokenUSDPrice(trxTokenAddress, priceBUSD)
	}

	valueUSD := calculateValueUSD(tokenAmount, trxAmount, priceAUSD, priceBUSD)

	syncEvent := models.LiquidityEvent{
		BlockNumber: p.state.Block.Number.Uint64(),
		Date:        time.Unix(timestamp, 0),
		Tx:          tx,
		Pair:        pair.ToBase58(),
		Chain:       Chain,
		Klass:       "sync",
		Wallet:      tronApi.FromHex(operator).ToBase58(),
		Order:       0,
		Reserve0:    tokenAmountRaw.String(),
		Reserve1:    trxAmountRaw.String(),
		Price:       priceA,
		PriceUSD:    priceAUSD,
		ReserveUSD:  valueUSD,
	}
	p.state.AddLiquidity(&syncEvent)
}

func (p *Parser) isPairWhiteListed(pair *tronApi.Address) bool {
	pairAddress := pair.ToBase58()
	if _, ok := p.whiteListedPairs.Load(pairAddress); ok {
		return true
	}
	return false
}

// onPairCreated - handle listing event
// topics - exchange, token
func (p *Parser) onPairCreated(log tronApi.Log, tx string, timestamp int64) {
	factory := tronApi.FromHex(log.Address)
	pair := tronApi.FromHex(tronApi.TrimZeroes(log.Topics[2]))
	//token := log.Topics[1]
	nodeUrl := os.Getenv("SOLIDITY_FULL_NODE_URL")
	p.state.RegisterNewPair(factory.ToBase58(), pair.ToBase58(), "sunswap", Chain, nodeUrl, time.Unix(timestamp, 0))
}

// convert "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" -> 0xddf252ad
func getMethodId(input string) int {
	// take first 8 symbols from string
	numberStr := input[0:8]
	// Convert into number
	n := new(big.Int)
	n.SetString(numberStr, 16)
	return int(n.Int64())
}
