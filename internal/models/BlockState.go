package models

import commonModels "github.com/kattana-io/models/pkg/storage"

type BlockState struct {
	DirectSwaps []*commonModels.DirectSwap     `json:"direct_swaps"`
	PairSwaps   []*commonModels.PairSwap       `json:"pair_swaps"`
	Liquidities []*commonModels.LiquidityEvent `json:"liquidity_events"`
	Transfers   []*commonModels.TransferEvent  `json:"transfer_events"`
	Pairs       []*NewPair                     `json:"new_pairs"`
	Block       *Block                         `json:"block"`
}
