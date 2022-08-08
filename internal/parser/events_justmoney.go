package parser

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

const JMFactoryABI = `[{"inputs":[{"internalType":"address","name":"_owner","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"_moderator","type":"address"}],"name":"addModerator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"allPairs","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"allPairsLength","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"}],"name":"createPair","outputs":[{"internalType":"address","name":"pair","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"feeTo","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"address","name":"","type":"address"}],"name":"getPair","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"pair","type":"address"}],"name":"getPairSymbols","outputs":[{"internalType":"string","name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getRouterAddress","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"getTokensByPair","outputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_moderator","type":"address"}],"name":"removeModerator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_feeTo","type":"address"}],"name":"setFeeTo","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_newOwner","type":"address"}],"name":"setNewOwner","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_router","type":"address"}],"name":"setRouterAddress","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

func (p *Parser) onJmPairCreated(log tronApi.Log, tx string, timestamp int64) {
	factory := tronApi.FromHex(log.Address)

	factoryAbi, err := abi.JSON(strings.NewReader(JMFactoryABI))
	if err != nil {
		p.log.Warn("Could not parse factory abi: ", zap.Error(err))
		return
	}

	event, err := factoryAbi.EventByID(common.HexToHash(log.Topics[0]))
	if event != nil {
		if err != nil {
			p.log.Debug("Unpack error", zap.Error(err))
			return
		}
		data := make(map[string]interface{})
		err2 := event.Inputs.UnpackIntoMap(data, common.FromHex(log.Data))
		if err2 != nil {
			p.log.Debug("Unpack error", zap.Error(err2))
			return
		}
		pairAddress := data["pair"].(common.Address)
		pair := tronApi.FromHex(pairAddress.Hex())
		nodeUrl := os.Getenv("SOLIDITY_FULL_NODE_URL")
		p.state.RegisterNewPair(factory.ToBase58(), pair.ToBase58(), "justmoney", Chain, nodeUrl, time.Unix(timestamp, 0))
	}
}
