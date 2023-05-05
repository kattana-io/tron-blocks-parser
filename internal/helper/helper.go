package helper

import (
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/shopspring/decimal"
	"math/big"
)

func TronValueToDecimal(data string) *decimal.Decimal {
	val := big.Int{}
	val.SetString(tronApi.TrimZeroes(data), models.TronBase)
	rowAmount := decimal.NewFromBigInt(&val, 0)

	return &rowAmount
}
