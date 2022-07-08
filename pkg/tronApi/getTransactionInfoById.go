package tronApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GetTransactionInfoByIdResp struct {
	Id              string   `json:"id"`
	Fee             int      `json:"fee"`
	BlockNumber     int      `json:"blockNumber"`
	BlockTimeStamp  int64    `json:"blockTimeStamp"`
	ContractResult  []string `json:"contractResult"`
	ContractAddress string   `json:"contract_address"`
	Receipt         struct {
		OriginEnergyUsage int    `json:"origin_energy_usage"`
		EnergyUsageTotal  int    `json:"energy_usage_total"`
		NetFee            int    `json:"net_fee"`
		Result            string `json:"result"`
	} `json:"receipt"`
	Log []Log `json:"log"`
}

type Log struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

func (a *Api) GetTransactionInfoById(id string) (*GetTransactionInfoByIdResp, error) {
	postBody, _ := json.Marshal(map[string]interface{}{
		"value": id,
	})
	responseBody := bytes.NewBuffer(postBody)

	res, err := http.Post(fmt.Sprintf("%s/walletsolidity/gettransactioninfobyid", a.endpoint), "application/json", responseBody)
	defer res.Body.Close()

	if err != nil {
		a.log.Error(err.Error())
		return &GetTransactionInfoByIdResp{}, err
	}

	var data GetTransactionInfoByIdResp
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)

	return &data, nil
}
