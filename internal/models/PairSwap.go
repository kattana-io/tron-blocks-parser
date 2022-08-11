package models

import (
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

type PairSwap struct {
	Tx          string          `json:"tx"`
	Date        time.Time       `json:"date"`
	Chain       string          `json:"chain"`
	BlockNumber uint64          `json:"blocknumber"`
	Pair        string          `json:"pair"`
	Amount0     *big.Int        `json:"amount0"`
	Amount1     *big.Int        `json:"amount1"`
	Buy         bool            `json:"buy"`
	PriceA      decimal.Decimal `json:"pricea"`
	PriceAUSD   decimal.Decimal `json:"pricea_usd"`
	PriceB      decimal.Decimal `json:"priceb"`
	PriceBUSD   decimal.Decimal `json:"priceb_usd"`
	Bot         bool            `json:"bot"`
	Wallet      string          `json:"wallet"`
	Order       uint16          `json:"order"`
	ValueUSD    decimal.Decimal `json:"value_usd"`
}
