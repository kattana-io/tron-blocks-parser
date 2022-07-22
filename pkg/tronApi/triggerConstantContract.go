package tronApi

import (
	"github.com/goccy/go-json"
	"strconv"
)

const DummyCaller = "410000000000000000000000000000000000000000"

type TCCResponse struct {
	Result struct {
		Result bool `json:"result"`
	} `json:"result"`
	EnergyUsed     int      `json:"energy_used"`
	ConstantResult []string `json:"constant_result"`
	Transaction    struct {
		Ret []struct {
		} `json:"ret"`
		Visible bool   `json:"visible"`
		TxID    string `json:"txID"`
		RawData struct {
			Contract []struct {
				Parameter struct {
					Value struct {
						Data            string `json:"data"`
						OwnerAddress    string `json:"owner_address"`
						ContractAddress string `json:"contract_address"`
					} `json:"value"`
					TypeUrl string `json:"type_url"`
				} `json:"parameter"`
				Type string `json:"type"`
			} `json:"contract"`
			RefBlockBytes string `json:"ref_block_bytes"`
			RefBlockHash  string `json:"ref_block_hash"`
			Expiration    int64  `json:"expiration"`
			Timestamp     int64  `json:"timestamp"`
		} `json:"raw_data"`
		RawDataHex string `json:"raw_data_hex"`
	} `json:"transaction"`
}

// TriggerConstantContract /**
func (a *Api) TriggerConstantContract(contractAddress string, functionSelector string, parameter string) (*TCCResponse, error) {
	postBody, _ := json.Marshal(map[string]interface{}{
		"owner_address":     DummyCaller,
		"contract_address":  contractAddress,
		"function_selector": functionSelector,
		"parameter":         parameter,
	})

	res, err := a.provider.Request(a.provider.TriggerConstantContract(), postBody)
	defer res.Body.Close()

	if err != nil {
		a.log.Error(err.Error())
		return &TCCResponse{}, err
	}

	var data TCCResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return &TCCResponse{}, err
	}

	return &data, nil
}

func (a *Api) GetTokenDecimals(token string) (int32, error) {
	nA := normalizeAddress(token)
	postBody, _ := json.Marshal(map[string]interface{}{
		"owner_address":     DummyCaller,
		"contract_address":  nA,
		"function_selector": "decimals()",
	})

	res, err := a.provider.Request(a.provider.TriggerConstantContract(), postBody)
	defer res.Body.Close()

	if err != nil {
		a.log.Error(err.Error())
		return 18, err
	}

	var data TCCResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil || len(data.ConstantResult) == 0 {
		return 18, err
	}

	result, err := strconv.ParseInt(TrimZeroes(data.ConstantResult[0]), 16, 16)
	if err != nil {
		return 18, err
	}
	return int32(result), nil
}

func (a *Api) GetPairToken(pair string) (string, error) {
	nA := normalizeAddress(pair)
	postBody, _ := json.Marshal(map[string]interface{}{
		"owner_address":     DummyCaller,
		"contract_address":  nA,
		"function_selector": "tokenAddress()",
	})

	res, err := a.provider.Request(a.provider.TriggerConstantContract(), postBody)
	defer res.Body.Close()

	if err != nil {
		a.log.Error(err.Error())
		return "", err
	}

	var data TCCResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil || len(data.ConstantResult) == 0 {
		return "", err
	}

	return TrimZeroes(data.ConstantResult[0]), nil
}
