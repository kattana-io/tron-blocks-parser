package models

import "math/big"

type QuotePair struct {
	Title   string   `json:"Title"`
	Token   string   `json:"Token"`
	Kind    uint8    `json:"Kind"`
	Flipped bool     `json:"Flipped"`
	Before  *big.Int `json:"Before,omitempty"`
	After   *big.Int `json:"After,omitempty"`
}
