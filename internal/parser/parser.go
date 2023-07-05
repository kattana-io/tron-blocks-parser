package parser

import (
	"fmt"
	"sync"

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
	api              *tronApi.API
	log              *zap.Logger
	failedTx         []tronApi.Transaction
	txMap            map[string]*tronApi.Transaction
	state            *State
	tokenLists       *integrations.TokenListsProvider
	pairsCache       *cache.PairsCache
	fiatConverter    *converters.FiatConverter
	abiHolder        *abi.Holder
	jmcache          *cache.JMPairsCache
	whiteListedPairs *sync.Map
}

// Parse - parse single block
func (p *Parser) Parse(block models.Block) bool {
	p.state = CreateState(&block)

	resp, err := p.api.GetBlockByNum(int32(block.Number.Int64()))
	if resp.BlockID == "" {
		p.log.Sugar().Errorf("Could not receive block: ", zap.Error(err))
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
		p.txMap[resp.Transactions[i].TxID] = &resp.Transactions[i]
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

func (p *Parser) parseTransactions(blockNumber int64) {
	resp, err := p.api.GetTransactionInfoByBlockNum(blockNumber)

	if err != nil {
		p.log.Error("parseTransaction: " + err.Error())
		return
	}
	wg := sync.WaitGroup{}

	for _, tx := range resp {
		if tx.Receipt.Result != "SUCCESS" {
			continue
		}
		wg.Add(len(tx.Log))

		// Process logs
		for _, log := range tx.Log {
			t := tx.BlockTimeStamp / 1000
			owner := p.txMap[tx.ID].RawData.Contract[0].Parameter.Value.OwnerAddress

			go p.processLog(log, tx.ID, t, owner, &wg)
		}
	}
	wg.Wait()
}

func (p *Parser) GetEncodedBlock() []byte {
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

func New(api *tronApi.API,
	log *zap.Logger,
	lists *integrations.TokenListsProvider,
	pairsCache *cache.PairsCache,
	converter *converters.FiatConverter,
	abiHolder *abi.Holder,
	jmCache *cache.JMPairsCache) *Parser {
	whiteListedPairs := sync.Map{}
	whiteListedPairs.Store("TFGDbUyP8xez44C76fin3bn3Ss6jugoUwJ", true) // TRX-USDT v2
	whiteListedPairs.Store("TNLcz8A9hGKbTNJ6b6C1GTyigwxURbWzkM", true) // USDD-USDT
	whiteListedPairs.Store("TQcia2H2TU3WrFk9sKtdK9qCfkW8XirfPQ", true) // TRX-USDJ
	whiteListedPairs.Store("TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE", true) // USDT-TRX
	whiteListedPairs.Store("TXX1i3BWKBuTxUmTERCztGyxSSpRagEcjX", true) // USDC-TRX
	whiteListedPairs.Store("TSJWbBJAS8HgQCMJfY5drVwYDa7JBAm6Es", true) // USDD-TRX
	whiteListedPairs.Store("TYukBQZ2XXCcRCReAUguyXncCWNY9CEiDQ", true) // JST-TRX

	return &Parser{
		jmcache:          jmCache,
		fiatConverter:    converter,
		api:              api,
		log:              log,
		failedTx:         []tronApi.Transaction{},
		txMap:            make(map[string]*tronApi.Transaction),
		tokenLists:       lists,
		pairsCache:       pairsCache,
		abiHolder:        abiHolder,
		whiteListedPairs: &whiteListedPairs,
	}
}
