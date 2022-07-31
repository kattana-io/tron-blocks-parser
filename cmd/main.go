package main

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/abi"
	"github.com/kattana-io/tron-blocks-parser/internal/cache"
	"github.com/kattana-io/tron-blocks-parser/internal/converters"
	"github.com/kattana-io/tron-blocks-parser/internal/integrations"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"github.com/kattana-io/tron-blocks-parser/internal/parser"
	"github.com/kattana-io/tron-blocks-parser/internal/runway"
	"github.com/kattana-io/tron-blocks-parser/internal/transport"
	"github.com/kattana-io/tron-blocks-parser/pkg/tronApi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

func main() {
	/**
	 * Handle run options like close signals
	 */
	runner := runway.Create()
	logger := runner.Logger()
	abiHolder := abi.Create(logger)
	registerCommandLineFlags(logger)
	mode, topic := getRunningMode()
	redis := runner.Redis()
	runner.Run()

	api := createApi(logger)
	tokenLists := integrations.NewTokensListProvider(logger)
	pairsCache := cache.NewPairsCache(redis, logger)
	shouldWarmupCache(pairsCache)

	logger.Info(fmt.Sprintf("Start parser in %s mode", mode))
	publisher := transport.NewPublisher("parser.sys.parsed", os.Getenv("KAFKA"), logger)
	t := transport.CreateConsumer(topic, logger)
	t.OnBlock(func(Value []byte) bool {
		/**
		 * Decode block
		 */
		block := models.Block{}
		err := json.Unmarshal(Value, &block)
		if err != nil {
			logger.Error(err.Error())
			return false
		}
		/**
		 * Check for valid block number
		 * Why we can have 0 here? Invalid value in JSON
		 */
		if block.Number.Int64() == 0 {
			logger.Info("Received null block number, skipping")
			return true
		}
		/**
		 * Process block
		 */
		fiatConverter := converters.CreateConverter(redis, logger, &block)
		p := parser.New(api, logger, tokenLists, pairsCache, fiatConverter, abiHolder)
		ok := p.Parse(block)
		if ok {
			publisher.PublishBlock(context.Background(), p.GetEncodedBlock())
			return true
		} else {
			return publisher.PublishFailedBlock(context.Background(), block)
		}
	})

	t.OnFail(func(err error) {
		logger.Fatal("failed to close reader: " + err.Error())
	})
	t.Listen()
}

// Check if we should fill cache for dev purpose
func shouldWarmupCache(pairsCache *cache.PairsCache) {
	warmupFlag := os.Getenv("PAIRS_WARMUP")
	warmup := warmupFlag == "true"

	if warmup {
		ssa := integrations.NewSunswapStatisticsAdapter()
		if ssa.Ok {
			pairsCache.Warmup(ssa.TokenPairs)
		}
	}
}

func createApi(logger *zap.Logger) *tronApi.Api {
	nodeUrl := os.Getenv("SOLIDITY_FULL_NODE_URL")
	var provider tronApi.ApiUrlProvider
	if nodeUrl == "" {
		logger.Info("Using trongrid adapter")
		provider = tronApi.NewTrongridUrlProvider()
	} else {
		logger.Info("Using node adapter")
		provider = tronApi.NewNodeUrlProvider(nodeUrl)
	}
	api := tronApi.NewApi(nodeUrl, logger, provider)
	return api
}

//getRunningMode - decide should we consumer live or history
func getRunningMode() (mode string, topic models.Topics) {
	mode = viper.GetString("mode")

	if mode == "HISTORY" {
		topic = models.TRON_HISTORY
	} else {
		topic = models.TRON_LIVE
	}
	return mode, topic
}

func registerCommandLineFlags(logger *zap.Logger) {
	rootCmd := &cobra.Command{}
	rootCmd.Flags().String("mode", string(models.LIVE), "Please provide mode: --mode LIVE or --mode HISTORY")

	err := viper.BindPFlag("mode", rootCmd.Flags().Lookup("mode"))
	if err != nil {
		logger.Fatal(err.Error())
	}

	err = rootCmd.Execute()
	if err != nil {
		logger.Fatal(err.Error())
	}
}
