package tronApi

import "fmt"

type ApiUrlProvider interface {
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

func NewNodeUrlProvider(host string) ApiUrlProvider {
	return &NodeUrlProvider{
		host: host,
	}
}

/**
 * TronGrid URLs
 */

func NewTrongridUrlProvider() ApiUrlProvider {
	return &TrongridUrlProvider{}
}

type TrongridUrlProvider struct{}

func (n TrongridUrlProvider) GetBlockByNum() string {
	return "https://api.trongrid.io/wallet/getblockbynum"
}

func (n TrongridUrlProvider) GetTransactionInfoById() string {
	return "https://api.trongrid.io/wallet/gettransactioninfobyid"
}

func (n TrongridUrlProvider) TriggerConstantContract() string {
	return "https://api.trongrid.io/wallet/triggerconstantcontract"
}
