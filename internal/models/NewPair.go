package models

type NewPair struct {
	Factory     string `json:"factory"`
	Pair        string `json:"pair"`
	Klass       string `json:"klass"`
	Network     string `json:"network"`
	Node        string `json:"node"`
	PoolCreated int64  `json:"pool_created"`
}
