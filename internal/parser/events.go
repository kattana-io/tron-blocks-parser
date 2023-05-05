package parser

/**
 * Package with event handling logic
 */

import (
	commonModels "github.com/kattana-io/models/pkg/storage"
	"github.com/kattana-io/tron-blocks-parser/internal/helper"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"sync"
	"time"
)

const (
	Chain           = "TRON"
	SyncTopicsCount = 4
)

/**
 * List of supported events
 */
const transferEvent = 0xddf252ad

// const liquidityAdded = 0x06239653
// const liquidityRemoved = 0x0fbf06c0
const tokenPurchaseEvent = 0xcd60aa75
const trxPurchaseEvent = 0xdad9ec5c
const snapshotEvent = 0xcc7244d3
const listingEvent = 0x9d42cb01
const jmListingEvent = 0x0d3648bd
const jmUniv2SwapEvent = 0xd78ad95f
const jmUniV2SyncEventID = 0x1c411e9a // 0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1

func isBase58(input string) bool {
	return input[0] == 'T'
}

func getAddressObject(pair string) *tronApi.Address {
	if isBase58(pair) {
		return tronApi.FromBase58(pair)
	}
	return tronApi.FromHex(pair)
}

func (p *Parser) processLog(log tronApi.Log, tx string, timestamp int64, owner string, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(log.Topics) < 1 {
		return
	}
	methodID := getMethodID(log.Topics[0])

	ownerAddress := getAddressObject(owner)
	switch methodID {
	case transferEvent:
		p.processHolder(log, tx)
	case tokenPurchaseEvent:
		p.onTokenPurchase(log, tx, timestamp)
	case trxPurchaseEvent:
		p.onTrxPurchase(log, tx, timestamp)
	case snapshotEvent:
		p.onPairSnapshot(log, tx, timestamp)
	case listingEvent:
		p.onPairCreated(log, timestamp)
	case jmListingEvent:
		p.onJmPairCreated(log, timestamp)
	case jmUniv2SwapEvent:
		p.onJmSwapEvent(log, tx, ownerAddress, timestamp)
	case jmUniV2SyncEventID:
		p.onJmSyncEvent(log, tx, ownerAddress, timestamp)
	}
}

func (p *Parser) processHolder(log tronApi.Log, tx string) {
	amounts, ok := big.NewInt(0).SetString(log.Data, models.TronBase)
	if !ok || amounts.Cmp(big.NewInt(0)) == 0 {
		return
	}

	token := tronApi.FromHex(log.Address).ToBase58()
	from := tronApi.FromHex(tronApi.TrimZeroes(log.Topics[1])).ToBase58()
	to := tronApi.FromHex(tronApi.TrimZeroes(log.Topics[2])).ToBase58()

	h := commonModels.Holder{
		Token: token,
		From:  from,
		To:    to,
		Tx:    tx,
	}
	p.state.AddProcessHolder(&h)
}

func (p *Parser) parseTransferContract(transaction *tronApi.Transaction) {
	h := commonModels.Holder{
		Token:  models.NativeToken,
		From:   tronApi.FromHex(tronApi.TrimZeroes(transaction.RawData.Contract[0].Parameter.Value.OwnerAddress)).ToBase58(),
		To:     tronApi.FromHex(tronApi.TrimZeroes(transaction.RawData.Contract[0].Parameter.Value.ToAddress)).ToBase58(),
		Tx:     transaction.TxID,
		Amount: transaction.RawData.Contract[0].Parameter.Value.Amount,
	}
	p.state.AddProcessHolder(&h)
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

	swap := commonModels.PairSwap{
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

	swap := commonModels.PairSwap{
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

func calculateValueUSD(amount0, amount1, ausd, busd decimal.Decimal) decimal.Decimal {
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
	if len(log.Topics) != SyncTopicsCount {
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
	priceB := tokenAmount.Div(trxAmount)

	priceAUSD, priceBUSD := p.fiatConverter.ConvertAB(tokenA, tokenB, priceA)
	p.fiatConverter.UpdateTokenUSDPrice(tokenA, priceAUSD)

	if p.isPairWhiteListed(pair) {
		p.fiatConverter.UpdateTokenUSDPrice(trxTokenAddress, priceBUSD)
	}

	valueUSD := calculateValueUSD(tokenAmount, trxAmount, priceAUSD, priceBUSD)

	syncEvent := commonModels.LiquidityEvent{
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
		PriceA:      priceA,
		PriceAUSD:   priceAUSD,
		PriceB:      priceB,
		PriceBUSD:   priceBUSD,
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
func (p *Parser) onPairCreated(log tronApi.Log, timestamp int64) {
	factory := tronApi.FromHex(log.Address)
	pair := tronApi.FromHex(tronApi.TrimZeroes(log.Topics[2]))
	nodeURL := os.Getenv("SOLIDITY_FULL_NODE_URL")
	p.state.RegisterNewPair(factory.ToBase58(), pair.ToBase58(), "sunswap", Chain, nodeURL, time.Unix(timestamp, 0))
}

// convert "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" -> 0xddf252ad
func getMethodID(input string) int {
	// take first 8 symbols from string
	numberStr := input[0:8]
	// Convert into number
	n := new(big.Int)
	n.SetString(numberStr, models.TronBase)
	return int(n.Int64())
}
