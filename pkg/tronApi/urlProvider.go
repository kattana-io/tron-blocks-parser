package tronApi

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

type ApiUrlProvider interface {
	Request(url string, body []byte) (*http.Response, error)
	GetBlockByNum() string
	GetTransactionInfoById() string
	TriggerConstantContract() string
}

type NodeUrlProvider struct {
	host string
}

func (n NodeUrlProvider) GetBlockByNum() string {
	return fmt.Sprintf("%s/walletsolidity/getblockbynum", n.host)
}

func (n NodeUrlProvider) GetTransactionInfoById() string {
	return fmt.Sprintf("%s/walletsolidity/gettransactioninfobyid", n.host)
}

func (n NodeUrlProvider) TriggerConstantContract() string {
	return fmt.Sprintf("%s/walletsolidity/triggerconstantcontract", n.host)
}

func (n NodeUrlProvider) Request(url string, body []byte) (*http.Response, error) {
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	return res, err
}

func NewNodeUrlProvider(host string) ApiUrlProvider {
	return &NodeUrlProvider{
		host: host,
	}
}

/**
 * TronGrid URLs
 */

func NewTrongridUrlProvider() ApiUrlProvider {
	return &TrongridUrlProvider{
		ApiKey: os.Getenv("TRONGRID_API_KEY"),
	}
}

type TrongridUrlProvider struct {
	ApiKey string
}

// Request - Add headers to request https://developers.tron.network/reference/api-key#how-to-use-api-keys
func (n TrongridUrlProvider) Request(url string, body []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if n.ApiKey != "" {
		req.Header.Add("TRON-PRO-API-KEY", n.ApiKey)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	return resp, err
}

func (n TrongridUrlProvider) GetBlockByNum() string {
	return "https://api.trongrid.io/wallet/getblockbynum"
}

func (n TrongridUrlProvider) GetTransactionInfoById() string {
	return "https://api.trongrid.io/wallet/gettransactioninfobyid"
}

func (n TrongridUrlProvider) TriggerConstantContract() string {
	return "https://api.trongrid.io/wallet/triggerconstantcontract"
}
