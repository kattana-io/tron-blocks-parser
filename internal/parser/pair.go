package parser

type Token struct {
	Address  string
	Decimals int
}

type Pair struct {
	address string
	token   Token
}

func (p *Pair) GetToken() Token {
	// call view function of pair: tokenAddress
	return Token{}
}

func (p *Pair) Dissolve() {

}

// @TODO: Cache it in redis
func (p *Pair) GetTokenAddress() string {
	return "TN3W4H6rK2ce4vX9YnFQHwKENnHjoxb3m9"
}
