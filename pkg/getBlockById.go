package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GetBlockByIdResp struct {
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

func (a *Api) GetBlockById(id string) *GetBlockByIdResp {
	postBody, _ := json.Marshal(map[string]interface{}{
		"value": id,
	})
	responseBody := bytes.NewBuffer(postBody)

	res, err := http.Post(fmt.Sprintf("%s/wallet/getblockbyid", a.endpoint), "application/json", responseBody)
	defer res.Body.Close()

	if err != nil {
		a.log.Error(err.Error())
		return &GetBlockByIdResp{}
	}

	var data GetBlockByIdResp
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)

	return &data
}
