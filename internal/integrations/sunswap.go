package integrations

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"go.uber.org/zap"
	"os"
)

type pairToken struct {
	Address  string `json:"address"`
	Decimals int32  `json:"decimals"`
}

func loadSunswapMapping() (map[string]pairToken, error) {
	raw, err := os.ReadFile("sunswap.json")
	if err != nil {
		return nil, err
	}
	data := make(map[string]pairToken)
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return data, err
}

type SunswapProvider struct {
	ok   bool
	list map[string]pairToken
}

func (t *SunswapProvider) GetToken(address string) (models.Token, bool) {
	val, ok := t.list[address]
	if ok {
		return models.Token{
			Address:  val.Address,
			Decimals: val.Decimals,
		}, true
	}
	return models.Token{}, false
}

func NewSunswapProvider() *SunswapProvider {
	list, err := loadSunswapMapping()
	ok := err == nil
	if ok {
		zap.L().Info(fmt.Sprintf("Loaded %d pairs", len(list)))
	}

	return &SunswapProvider{
		ok:   ok,
		list: list,
	}
}
