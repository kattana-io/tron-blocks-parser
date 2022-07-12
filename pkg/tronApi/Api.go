package tronApi

import (
	"encoding/hex"
	"github.com/mr-tron/base58"
	"go.uber.org/zap"
)

type Api struct {
	endpoint string
	log      *zap.Logger
}

func NewApi(nodeUrl string, logger *zap.Logger) *Api {
	return &Api{
		endpoint: nodeUrl,
		log:      logger,
	}
}

// Convert "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE" -> "41a2726afbecbd8e936000ed684cef5e2f5cf43008"
func normalizeAddress(address string) string {
	num, err := base58.Decode(address)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(num)[0:42]
}

// "41a2726afbecbd8e936000ed684cef5e2f5cf43008" -> "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE"
func decodeAddress(raw string) string {
	bts, err := hex.DecodeString(raw)
	if err != nil {
		return ""
	}
	return base58.Encode(bts) // todo: something wrong here
}

// "000000000000000000000000a614f803b6fd780986a42c78ec9c7f77e6ded13c" -> "a614f803b6fd780986a42c78ec9c7f77e6ded13c"
func trimZeroes(address string) string {
	idx := 0
	for ; idx < len(address); idx++ {
		if address[idx] != '0' {
			break
		}
	}
	return address[idx:]
}
