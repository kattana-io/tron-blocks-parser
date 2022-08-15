package parser

import (
	commonModels "github.com/kattana-io/models/pkg/storage"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"sync"
	"time"
)

type State struct {
	DirectSwaps     []*commonModels.DirectSwap     `json:"direct_swaps"`
	PairSwaps       []*commonModels.PairSwap       `json:"pair_swaps"`
	Liquidities     []*commonModels.LiquidityEvent `json:"liquidity_events"`
	Transfers       []*commonModels.TransferEvent  `json:"transfer_events"`
	Pairs           []*models.NewPair              `json:"new_pairs"`
	Block           *models.Block                  `json:"block"`
	pairsLock       *sync.Mutex
	tradesLock      *sync.Mutex
	liquidityLock   *sync.Mutex
	transfersLock   *sync.Mutex
	directSwapsLock *sync.Mutex
}

func CreateState(block *models.Block) *State {
	return &State{
		Block:           block,
		pairsLock:       &sync.Mutex{},
		transfersLock:   &sync.Mutex{},
		liquidityLock:   &sync.Mutex{},
		tradesLock:      &sync.Mutex{},
		directSwapsLock: &sync.Mutex{},
	}
}

func (i *State) AddTrade(trade *commonModels.PairSwap) {
	i.tradesLock.Lock()
	defer i.tradesLock.Unlock()
	i.PairSwaps = append(i.PairSwaps, trade)
}

func (i *State) AddLiquidity(liquidity *commonModels.LiquidityEvent) {
	i.liquidityLock.Lock()
	defer i.liquidityLock.Unlock()
	i.Liquidities = append(i.Liquidities, liquidity)
}

func (i *State) AddTransferEvent(transfer *commonModels.TransferEvent) {
	i.transfersLock.Lock()
	defer i.transfersLock.Unlock()
	i.Transfers = append(i.Transfers, transfer)
}

func (i *State) RegisterNewPair(Factory string, Pair string, Klass string, network string, node string, time time.Time) {
	i.pairsLock.Lock()
	defer i.pairsLock.Unlock()

	i.Pairs = append(i.Pairs, &models.NewPair{
		Factory:     Factory,
		Pair:        Pair,
		Klass:       Klass,
		Network:     network,
		Node:        node,
		PoolCreated: time.Unix(),
	})
}

func (i *State) AddDirectSwap(m *commonModels.DirectSwap) {
	i.directSwapsLock.Lock()
	defer i.directSwapsLock.Unlock()
	i.DirectSwaps = append(i.DirectSwaps, m)
}
