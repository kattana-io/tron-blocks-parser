package tronApi

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/mr-tron/base58"
	"go.uber.org/zap"
)

type Api struct {
	endpoint string
	log      *zap.Logger
	provider ApiUrlProvider
}

func NewApi(nodeUrl string, logger *zap.Logger, provider ApiUrlProvider) *Api {
	return &Api{
		endpoint: nodeUrl,
		log:      logger,
		provider: provider,
	}
}

// Convert "fba3416f7aac8ea9e12b950914d592c15c884372" -> "41fba3416f7aac8ea9e12b950914d592c15c884372"
func normalizeAddress(address string) string {
	return "41" + address
}

func s256(s []byte) []byte {
	h := sha256.New()
	h.Write(s)
	bs := h.Sum(nil)
	return bs
}

// "41a2726afbecbd8e936000ed684cef5e2f5cf43008" -> "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE"
func DecodeAddress(raw string) string {
	address := normalizeAddress(raw)
	addb, _ := hex.DecodeString(address)
	hash1 := s256(s256(addb))
	secret := hash1[:4]
	for _, v := range secret {
		addb = append(addb, v)
	}

	res := base58.Encode(addb)

	return res
}

// "000000000000000000000000a614f803b6fd780986a42c78ec9c7f77e6ded13c" -> "a614f803b6fd780986a42c78ec9c7f77e6ded13c"
func TrimZeroes(address string) string {
	idx := 0
	for ; idx < len(address); idx++ {
		if address[idx] != '0' {
			break
		}
	}
	return address[idx:]
}
