package helper

import (
	"github.com/kattana-io/tron-blocks-parser/pkg/tronApi"
	"github.com/shopspring/decimal"
	"math/big"
)

func TronValueToDecimal(data string) *decimal.Decimal {
	val := big.Int{}
	val.SetString(tronApi.TrimZeroes(data), 16)
	rowAmount := decimal.NewFromBigInt(&val, 0)

	return &rowAmount
}
