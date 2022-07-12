package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type TransferEvent struct {
	Chain       string          `json:"chain"`
	Contract    string          `json:"contract"`
	BlockNumber uint64          `json:"blocknumber"`
	Date        time.Time       `json:"date"`
	Order       uint            `json:"order"` // log order
	Tx          string          `json:"tx"`
	From        string          `json:"from"`   // usually contract
	To          string          `json:"to"`     // usually wallet
	Amount      string          `json:"amount"` // bigint amount
	ValueUSD    decimal.Decimal `json:"valueUSD"`
}
