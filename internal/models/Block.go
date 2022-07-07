package models

import (
	"math/big"
)

type Block struct {
	Type      string   `json:"type"`
	Network   string   `json:"network"`
	Number    *big.Int `json:"number"`
	Node      string   `json:"node"`
	Notify    bool     `json:"notify"`
	Timestamp uint64   `json:"timestamp"`
}
