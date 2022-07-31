package integrations

import (
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/intermediate"
	"net/http"
)

/**
 * EXPERIMENTAL
 * Motivation?
 * this adapter created for dev purposes
 * to be used to populate cache from the same request sunswap statistics page
 * to make you fit into trongrid limitations when running dev/tests process locally
 */

type SunswapStatisticsAdapter struct {
	Ok         bool
	TokenPairs []intermediate.Pair
}

type SSAListItem struct {
	Ver          int    `json:"ver"`
	Address      string `json:"address"`
	Volume14D    string `json:"volume14d"`
	TokenSymbol  string `json:"tokenSymbol"`
	IsValid      int    `json:"isValid"`
	TokenName    string `json:"tokenName"`
	Volume24Hrs  string `json:"volume24hrs"`
	TxId         string `json:"txId"`
	Liquidity    string `json:"liquidity"`
	Volume7D     string `json:"volume7d"`
	Type         string `json:"type"`
	TokenAddress string `json:"tokenAddress"`
	TokenLogoUrl string `json:"tokenLogoUrl"`
	TokenDecimal int    `json:"tokenDecimal"`
	Id           int    `json:"id"`
	Fees24Hrs    string `json:"fees24hrs"`
}

type ssaResponse struct {
	Code int `json:"code"`
	Data struct {
		TotalPages int           `json:"totalPages"`
		TotalCount string        `json:"totalCount"`
		List       []SSAListItem `json:"list"`
	} `json:"data"`
	Message string `json:"message"`
}

// Idea is to load locally top-1000 pairs by liquidity to populate cache once
func (s *SunswapStatisticsAdapter) fetchSSA() (*ssaResponse, error) {
	url := "https://abc.ablesdxd.link/swap/v2/exchanges/scan?pageNo=1&orderBy=liquidity&desc=true&pageSize=1000"

	res, err := http.Get(url)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	data := ssaResponse{}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, err
}

func NewSunswapStatisticsAdapter() *SunswapStatisticsAdapter {
	adapter := SunswapStatisticsAdapter{
		TokenPairs: make([]intermediate.Pair, 0),
	}
	resp, err := adapter.fetchSSA()
	if err != nil {
		adapter.Ok = false
		return &adapter
	}
	for _, pair := range resp.Data.List {
		adapter.TokenPairs = append(adapter.TokenPairs, intermediate.Pair{
			Address: pair.Address,
			Token: intermediate.Token{
				Address:  pair.TokenAddress,
				Decimals: int32(pair.TokenDecimal),
			},
		})
	}
	adapter.Ok = true
	return &adapter
}
