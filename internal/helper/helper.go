package helper

import (
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/shopspring/decimal"
	"math/big"
)

const TronBase = 16

func TronValueToDecimal(data string) *decimal.Decimal {
	val := big.Int{}
	val.SetString(tronApi.TrimZeroes(data), TronBase)
	rowAmount := decimal.NewFromBigInt(&val, 0)

	return &rowAmount
}
