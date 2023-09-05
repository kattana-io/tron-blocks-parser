package models

type Token struct {
	Address  string `json:"address"`
	Decimals int32  `json:"decimals"`
}

type Pair struct {
	Address string `json:"address"`
	Klass   string `json:"klass"`
	Token0  Token  `json:"token0"`
	Token1  Token  `json:"token1"`
}
