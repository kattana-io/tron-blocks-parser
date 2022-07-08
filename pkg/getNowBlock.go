package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GetNowBlockResp struct {
	BlockID     string `json:"blockID"`
	BlockHeader struct {
		RawData struct {
			Number         int    `json:"number"`
			TxTrieRoot     string `json:"txTrieRoot"`
			WitnessAddress string `json:"witness_address"`
			ParentHash     string `json:"parentHash"`
			Version        int    `json:"version"`
			Timestamp      int64  `json:"timestamp"`
		} `json:"raw_data"`
		WitnessSignature string `json:"witness_signature"`
	} `json:"block_header"`
	Transactions []struct {
		Ret []struct {
			ContractRet string `json:"contractRet"`
		} `json:"ret"`
		Signature []string `json:"signature"`
		TxID      string   `json:"txID"`
		RawData   struct {
			Contract []struct {
				Parameter struct {
					Value struct {
						Data            string `json:"data,omitempty"`
						OwnerAddress    string `json:"owner_address"`
						ContractAddress string `json:"contract_address,omitempty"`
						CallValue       int64  `json:"call_value,omitempty"`
						Amount          int    `json:"amount,omitempty"`
						AssetName       string `json:"asset_name,omitempty"`
						ToAddress       string `json:"to_address,omitempty"`
					} `json:"value"`
					TypeUrl string `json:"type_url"`
				} `json:"parameter"`
				Type string `json:"type"`
			} `json:"contract"`
			RefBlockBytes string `json:"ref_block_bytes"`
			RefBlockHash  string `json:"ref_block_hash"`
			Expiration    int64  `json:"expiration"`
			FeeLimit      int    `json:"fee_limit,omitempty"`
			Timestamp     int64  `json:"timestamp"`
		} `json:"raw_data"`
		RawDataHex string `json:"raw_data_hex"`
	} `json:"transactions"`
}

func (a *Api) GetNowBlock() *GetNowBlockResp {
	res, err := http.Get(fmt.Sprintf("%s/wallet/getnowblock", a.endpoint))
	defer res.Body.Close()

	if err != nil {
		a.log.Error(err.Error())
		return &GetNowBlockResp{}
	}

	var data GetNowBlockResp
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)

	return &data
}
