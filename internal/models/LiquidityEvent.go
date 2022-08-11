package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type LiquidityEvent struct {
	// generic info
	BlockNumber uint64    `json:"blocknumber"`
	Date        time.Time `json:"date"`
	Tx          string    `json:"tx"`
	Pair        string    `json:"pair"`
	Chain       string    `json:"chain"`
	Klass       string    `json:"klass"` // mint, burn, sync
	Wallet      string    `json:"wallet"`
	Order       uint      `json:"order"`
	// Mint, Burn
	Amount0 string `json:"amount0"`
	Amount1 string `json:"amount1"`
	// Sync
	Reserve0 string `json:"reserve0"`
	Reserve1 string `json:"reserve1"`
	// Statistics part
	PriceA    decimal.Decimal `json:"pricea"`
	PriceAUSD decimal.Decimal `json:"pricea_usd"`
	PriceB    decimal.Decimal `json:"priceb"`
	PriceBUSD decimal.Decimal `json:"priceb_usd"`
	// USD value of operation
	ValueUSD decimal.Decimal `json:"value_usd"`
	// Reserves value in USD
	ReserveUSD decimal.Decimal `json:"reserve_usd"`
}
