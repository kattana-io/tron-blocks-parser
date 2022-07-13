package intermediate

type Token struct {
	Address  string `json:"address"`
	Decimals int32  `json:"decimals"`
}

type Pair struct {
	Address string `json:"address"`
	Token   Token  `json:"token"`
}

func (p *Pair) GetToken() Token {
	// call view function of pair: tokenAddress
	return Token{}
}

func (p *Pair) Dissolve() {

}

func (p *Pair) SetToken(address string, decimals int32) {
	p.Token = Token{
		Address:  address,
		Decimals: decimals,
	}
}
