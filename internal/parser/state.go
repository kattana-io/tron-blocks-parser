package parser

import (
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"sync"
	"time"
)

type State struct {
	DirectSwaps     []*models.DirectSwap     `json:"direct_swaps"`
	PairSwaps       []*models.PairSwap       `json:"pair_swaps"`
	Liquidities     []*models.LiquidityEvent `json:"liquidity_events"`
	Transfers       []*models.TransferEvent  `json:"transfer_events"`
	Pairs           []*models.NewPair        `json:"new_pairs"`
	Block           *models.Block            `json:"block"`
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

func (i *State) AddTrade(trade *models.PairSwap) {
	i.tradesLock.Lock()
	defer i.tradesLock.Unlock()
	i.PairSwaps = append(i.PairSwaps, trade)
}

func (i *State) AddLiquidity(liquidity *models.LiquidityEvent) {
	i.liquidityLock.Lock()
	defer i.liquidityLock.Unlock()
	i.Liquidities = append(i.Liquidities, liquidity)
}

func (i *State) AddTransferEvent(transfer *models.TransferEvent) {
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

func (i *State) AddDirectSwap(m *models.DirectSwap) {
	i.directSwapsLock.Lock()
	defer i.directSwapsLock.Unlock()
	i.DirectSwaps = append(i.DirectSwaps, m)
}
