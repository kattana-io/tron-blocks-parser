package models

type BlockState struct {
	DirectSwaps []*DirectSwap     `json:"direct_swaps"`
	PairSwaps   []*PairSwap       `json:"pair_swaps"`
	Liquidities []*LiquidityEvent `json:"liquidity_events"`
	Transfers   []*TransferEvent  `json:"transfer_events"`
	Pairs       []*NewPair        `json:"new_pairs"`
	Block       *Block            `json:"block"`
}
