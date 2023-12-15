package parser

import (
	"fmt"
	"sync"

	"github.com/kattana-io/tron-objects-api/pkg/trc20"

	models "github.com/kattana-io/models/pkg/storage"
	"github.com/kattana-io/tron-blocks-parser/internal/abi"
	"github.com/kattana-io/tron-blocks-parser/internal/cache"
	"github.com/kattana-io/tron-blocks-parser/internal/converters"
	"github.com/kattana-io/tron-blocks-parser/internal/integrations"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

type Parser struct {
	api           *tronApi.API
	failedTx      []tronApi.Transaction
	txMap         sync.Map
	state         *State
	pairsCache    cache.PairCache
	fiatConverter *converters.FiatConverter
	abiHolder     *abi.Holder
	tokenLists    *integrations.TokenListsProvider
	sunswapPairs  *integrations.SunswapProvider
	log           *zap.SugaredLogger
}

// Parse - parse single block
func (p *Parser) Parse(block models.Block) bool {
	p.state = CreateState(&block)

	resp, err := p.api.GetBlockByNum(int32(block.Number.Int64()))
	if resp.BlockID == "" {
		p.log.Error("could not receive block: ", zap.Error(err))
		return false
	}
	if err != nil {
		p.log.Error("Parse: " + err.Error())
		return false
	}

	p.log.Info("Parsing block: " + block.Number.String())

	cnt := 0
	for i := range resp.Transactions {
		cnt++
		p.txMap.Store(resp.Transactions[i].TxID, &resp.Transactions[i])
	}

	p.parseTransactions(block.Number.Int64())
	p.log.Info(fmt.Sprintf("Parsing transactions: %v", cnt))

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

// parseTransactions - downloads block transactions and logs
func (p *Parser) parseTransactions(blockNumber int64) {
	resp, err := p.api.GetTransactionInfoByBlockNum(blockNumber)

	if err != nil {
		p.log.Error("parseTransaction: " + err.Error())
		return
	}

	for _, tx := range resp {
		if tx.Receipt.Result != "SUCCESS" {
			continue
		}

		// Process logs
		for _, log := range tx.Log {
			t := tx.BlockTimeStamp / 1000
			txRaw, ok := p.txMap.Load(tx.ID)

			if !ok {
				continue
			}
			owner := txRaw.(*tronApi.Transaction).RawData.Contract[0].Parameter.Value.OwnerAddress

			p.processLog(log, tx.ID, t, owner)
		}
	}
}

func (p *Parser) GetEncodedBlock() []byte {
	p.state.Block.Timestamp /= 1000 // consumer service expect to get timestamp in seconds
	b, err := msgpack.Marshal(p.state)
	if err != nil {
		p.log.Warn("GetEncodedBlock: " + err.Error())
		return nil
	}
	return b
}

func (p *Parser) GetEncodedHolders() []byte {
	holdersBlock := &models.HoldersBlock{
		Block:     p.state.Block.Number.Uint64(),
		Timestamp: int64(p.state.Block.Timestamp),
		Chain:     p.state.Block.Network,
		Holders:   p.state.Holders,
		Notify:    p.state.Block.Notify,
	}
	b, err := msgpack.Marshal(holdersBlock)
	if err != nil {
		p.log.Error("PublishHolders: " + err.Error())
		return nil
	}
	return b
}

func (p *Parser) DeleteHolders() {
	p.state.Holders = nil
}

func (p *Parser) GetTokenDecimals(address *tronApi.Address) (int32, bool) {
	token := trc20.New(p.api, address)
	return token.TryToGetDecimals(3)
}

func New(api *tronApi.API,
	lists *integrations.TokenListsProvider,
	pairsCache cache.PairCache,
	converter *converters.FiatConverter,
	abiHolder *abi.Holder,
	swLists *integrations.SunswapProvider) *Parser {
	return &Parser{
		fiatConverter: converter,
		api:           api,
		log:           zap.L().Sugar(),
		failedTx:      []tronApi.Transaction{},
		txMap:         sync.Map{},
		tokenLists:    lists,
		pairsCache:    pairsCache,
		abiHolder:     abiHolder,
		sunswapPairs:  swLists,
	}
}
