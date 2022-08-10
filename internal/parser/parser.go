package parser

import (
	"context"
	"github.com/kattana-io/tron-blocks-parser/internal/abi"
	"github.com/kattana-io/tron-blocks-parser/internal/cache"
	"github.com/kattana-io/tron-blocks-parser/internal/converters"
	"github.com/kattana-io/tron-blocks-parser/internal/integrations"
	"github.com/kattana-io/tron-blocks-parser/internal/intermediate"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Parser struct {
	api           *tronApi.Api
	log           *zap.Logger
	failedTx      []tronApi.Transaction
	txMap         map[string]*tronApi.GetTransactionInfoByIdResp
	state         *State
	tokenLists    *integrations.TokenListsProvider
	pairsCache    *cache.PairsCache
	fiatConverter *converters.FiatConverter
	abiHolder     *abi.Holder
	jmcache       *cache.JMPairsCache
}

// Parse - parse single block
func (p *Parser) Parse(block models.Block) bool {
	p.state = CreateState(&block)

	resp, err := p.api.GetBlockByNum(int32(block.Number.Int64()))
	if resp.BlockID == "" {
		p.log.Error("Could not receive block")
		return false
	}
	if err != nil {
		p.log.Error("Parse: " + err.Error())
		return false
	}

	p.log.Info("Parsing block: " + block.Number.String())

	cnt := 0
	for _, transaction := range resp.Transactions {
		if (isSuccessCall(&transaction) || hasContractCalls(&transaction)) && isNotTransferCall(&transaction) {
			cnt += 1
			p.parseTransaction(transaction)
		}
	}

	if len(p.failedTx) > 0 {
		for _, tx := range p.failedTx {
			p.parseTransaction(tx)
		}
	}
	// save prices
	p.fiatConverter.Commit()
	return true
}

// hasContractCalls - trading events are always contract calls
func hasContractCalls(transaction *tronApi.Transaction) bool {
	return len(transaction.RawData.Contract) >= 1
}

// isNotTransferCall - We don't need transfer events for trading
func isNotTransferCall(transaction *tronApi.Transaction) bool {
	return transaction.RawData.Contract[0].Type != "TransferContract"
}

// isSuccessCall- Do not download failed transactions
func isSuccessCall(transaction *tronApi.Transaction) bool {
	if len(transaction.Ret) < 1 {
		return false
	}
	return transaction.Ret[0].ContractRet == "SUCCESS"
}

// parseTransaction - process single transactions
func (p *Parser) parseTransaction(transaction tronApi.Transaction) {
	// Fetch transaction with logs
	resp, err := p.api.GetTransactionInfoById(transaction.TxID)
	if err != nil {
		p.log.Error("parseTransaction: " + err.Error())
		p.failedTx = append(p.failedTx, transaction)
		return
	}
	// Populate cache
	p.txMap[transaction.TxID] = resp

	// Process logs
	wg := sync.WaitGroup{}
	wg.Add(len(resp.Log))
	for _, log := range resp.Log {
		t := transaction.RawData.Timestamp / 1000
		go p.processLog(log, transaction.TxID, t, &wg)
	}
	wg.Wait()
}

const trxTokenAddress = "TRX"
const trxDecimals = 6

// GetCachePairToken - Try to pull from cache or populate cache
func (p *Parser) GetCachePairToken(address string) (string, int32, bool) {
	pair, err := p.pairsCache.Value(context.Background(), address)
	if err != nil {
		pInstance := intermediate.Pair{Address: address}
		tokenAddress, err := p.api.GetPairToken(address)
		if err != nil {
			p.log.Error("GetCachePairToken: " + err.Error())
			return "", 0, false
		}
		// Check list
		decimals, ok := p.tokenLists.GetDecimals(tokenAddress)
		if ok {
			pInstance.SetToken(tokenAddress, decimals)
		} else {
			// Call API
			decimals, err := p.api.GetTokenDecimals(tokenAddress)
			if err != nil {
				p.log.Error("GetCachePairToken: GetTokenDecimals: " + err.Error())
				return "", 0, false
			}
			pInstance.SetToken(tokenAddress, decimals)
		}
		err = p.pairsCache.Store(context.Background(), address, pInstance, time.Hour*2)
		if err != nil {
			p.log.Error("Could not put into cache: " + err.Error())
		}
		return pInstance.Token.Address, pInstance.Token.Decimals, true
	}
	return pair.Token.Address, pair.Token.Decimals, true
}

// GetPairTokens - Get tokens of pair
func (p *Parser) GetPairTokens(address *tronApi.Address) (string, int32, string, int32, bool) {
	adr, decimals, ok := p.GetCachePairToken(address.ToBase58())
	if ok {
		return adr, decimals, trxTokenAddress, trxDecimals, true
	}

	// Cache miss
	hexTokenAddress, err := p.api.GetPairToken(address.ToHex())
	tokenAddr := tronApi.FromHex(hexTokenAddress)
	cachedDecimals, ok := p.tokenLists.GetDecimals(tokenAddr.ToBase58())
	if ok {
		return tokenAddr.ToBase58(), cachedDecimals, trxTokenAddress, trxDecimals, true
	}

	dec, err := p.api.GetTokenDecimals(tokenAddr.ToHex())
	if err != nil {
		p.log.Error("GetPairTokens: " + err.Error())
		return "", 0, trxTokenAddress, trxDecimals, false
	}
	return tokenAddr.ToBase58(), dec, trxTokenAddress, trxDecimals, true
}

func (p *Parser) GetUniv2PairTokens(address *tronApi.Address) (string, int32, string, int32, bool) {
	pair, ok := p.jmcache.GetPair(Chain, address)
	if ok {
		return pair.TokenA.Address, int32(pair.TokenA.Decimals), pair.TokenB.Address, int32(pair.TokenB.Decimals), true
	}

	return "", 0, "", 0, false
}

func (p *Parser) GetEncodedBlock() []byte {
	b, err := msgpack.Marshal(p.state)
	if err != nil {
		p.log.Warn("GetEncodedBlock: " + err.Error())
		return nil
	}
	return b
}

func New(api *tronApi.Api, log *zap.Logger, lists *integrations.TokenListsProvider, pairsCache *cache.PairsCache, converter *converters.FiatConverter, abiHolder *abi.Holder, jmcache *cache.JMPairsCache) *Parser {
	return &Parser{
		jmcache:       jmcache,
		fiatConverter: converter,
		api:           api,
		log:           log,
		failedTx:      []tronApi.Transaction{},
		txMap:         make(map[string]*tronApi.GetTransactionInfoByIdResp),
		tokenLists:    lists,
		pairsCache:    pairsCache,
		abiHolder:     abiHolder,
	}
}
