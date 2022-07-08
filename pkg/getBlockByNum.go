package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GetBlockByNumResp struct {
	BlockID     string `json:"blockID"`
	BlockHeader struct {
		RawData struct {
			Number         int    `json:"number"`
			TxTrieRoot     string `json:"txTrieRoot"`
			WitnessAddress string `json:"witness_address"`
			ParentHash     string `json:"parentHash"`
			Timestamp      int64  `json:"timestamp"`
		} `json:"raw_data"`
		WitnessSignature string `json:"witness_signature"`
	} `json:"block_header"`
}

func (a *Api) GetBlockByNum(number int32) *GetBlockByNumResp {
	postBody, _ := json.Marshal(map[string]interface{}{
		"num": number,
	})
	responseBody := bytes.NewBuffer(postBody)

	res, err := http.Post(fmt.Sprintf("%s/wallet/getblockbynum", a.endpoint), "application/json", responseBody)
	defer res.Body.Close()

	if err != nil {
		a.log.Error(err.Error())
		return &GetBlockByNumResp{}
	}

	var data GetBlockByNumResp
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)

	return &data
}
