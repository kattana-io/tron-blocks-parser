package models

import (
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

type DirectSwap struct {
	Tx          string          `json:"tx"`
	Date        time.Time       `json:"date"`
	Chain       string          `json:"chain"`
	BlockNumber uint64          `json:"blocknumber"`
	Protocol    string          `json:"protocol"`
	SrcToken    string          `json:"src_token"`
	DstToken    string          `json:"dst_token"`
	Amount0     *big.Int        `json:"amount0"`
	Amount1     *big.Int        `json:"amount1"`
	PriceA      decimal.Decimal `json:"pricea"`
	PriceAUSD   decimal.Decimal `json:"pricea_usd"`
	PriceB      decimal.Decimal `json:"priceb"`
	PriceBUSD   decimal.Decimal `json:"priceb_usd"`
	Wallet      string          `json:"wallet"`
	Order       uint16          `json:"order"`
	ValueUSD    decimal.Decimal `json:"value_usd"`
}
