package parser

import (
	"github.com/kattana-io/tron-blocks-parser/internal/integrations"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"github.com/kattana-io/tron-blocks-parser/pkg/tronApi"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"sync"
)

type Parser struct {
	api        *tronApi.Api
	log        *zap.Logger
	failedTx   []string
	txMap      map[string]*tronApi.GetTransactionInfoByIdResp
	state      *State
	tokenLists *integrations.TokenListsProvider
}

// Parse - parse single block
func (p *Parser) Parse(block models.Block) bool {
	p.state = CreateState(&block)

	resp, err := p.api.GetBlockByNum(int32(block.Number.Int64()))
	if err != nil {
		p.log.Error("Parse: " + err.Error())
		return false
	}

	p.log.Info("Parsing block: " + block.Number.String())

	for _, transaction := range resp.Transactions {
		p.parseTransaction(transaction)
	}
	return true
}

// parseTransaction - process single transactions
func (p *Parser) parseTransaction(transaction tronApi.Transaction) {
	// Fetch transaction with logs
	resp, err := p.api.GetTransactionInfoById(transaction.TxID)
	if err != nil {
		p.log.Error("parseTransaction: " + err.Error())
		p.failedTx = append(p.failedTx, transaction.TxID)
		return
	}
	// Populate cache
	p.txMap[transaction.TxID] = resp
	// Process logs
	wg := sync.WaitGroup{}
	wg.Add(len(resp.Log))
	for _, log := range resp.Log {
		go p.processLog(log, transaction.TxID, transaction.RawData.Timestamp, &wg)
	}
	wg.Wait()
}

const trxTokenAddress = "TRX"
const trxDecimals = 6

func (p *Parser) GetPairTokens(pair string) (string, int32, string, int32) {
	pInstance := Pair{address: pair}
	address := pInstance.GetTokenAddress()

	cachedDecimals, ok := p.tokenLists.GetDecimals(address)
	if ok {
		return address, cachedDecimals, trxTokenAddress, trxDecimals
	}
	dec, err := p.api.GetTokenDecimals(address)
	if err != nil {
		p.log.Error("GetPairTokens: " + err.Error())
		dec = 18
	}
	return address, dec, trxTokenAddress, trxDecimals
}

func (p *Parser) GetEncodedBlock() []byte {
	b, err := msgpack.Marshal(p.state)
	if err != nil {
		p.log.Warn("GetEncodedBlock: " + err.Error())
		return nil
	}
	return b
}

func New(api *tronApi.Api, log *zap.Logger, lists *integrations.TokenListsProvider) *Parser {
	return &Parser{
		api:        api,
		log:        log,
		failedTx:   []string{},
		txMap:      make(map[string]*tronApi.GetTransactionInfoByIdResp),
		tokenLists: lists,
	}
}
