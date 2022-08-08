package helper

import (
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/shopspring/decimal"
	"math/big"
)

func TronValueToDecimal(data string) *decimal.Decimal {
	val := big.Int{}
	val.SetString(tronApi.TrimZeroes(data), 16)
	rowAmount := decimal.NewFromBigInt(&val, 0)

	return &rowAmount
}
