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
	SrcToken    string          `json:"srcToken"`
	DstToken    string          `json:"dstToken"`
	Amount0     *big.Int        `json:"amount0"`
	Amount1     *big.Int        `json:"amount1"`
	PriceA      decimal.Decimal `json:"priceA"`
	PriceAUSD   decimal.Decimal `json:"PriceAUSD"`
	PriceB      decimal.Decimal `json:"priceB"`
	PriceBUSD   decimal.Decimal `json:"priceBUSD"`
	Wallet      string          `json:"wallet"`
	Order       uint16          `json:"order"`
	ValueUSD    decimal.Decimal `json:"valueUSD"`
}
