package main

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/abi"
	"github.com/kattana-io/tron-blocks-parser/internal/cache"
	"github.com/kattana-io/tron-blocks-parser/internal/converters"
	"github.com/kattana-io/tron-blocks-parser/internal/helper"
	"github.com/kattana-io/tron-blocks-parser/internal/integrations"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"github.com/kattana-io/tron-blocks-parser/internal/parser"
	"github.com/kattana-io/tron-blocks-parser/internal/runway"
	"github.com/kattana-io/tron-blocks-parser/internal/transport"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"github.com/kattana-io/tron-objects-api/pkg/url"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

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
	quotesFile := helper.NewQuotesFile()
	tokenLists := integrations.NewTokensListProvider(logger)
	pairsCache := cache.NewPairsCache(redis, logger)
	jmPairsCache := cache.CreateJMPairsCache(redis, api, tokenLists, logger)
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
		fiatConverter := converters.CreateConverter(redis, logger, &block, quotesFile.Get())
		p := parser.New(api, logger, tokenLists, pairsCache, fiatConverter, abiHolder, jmPairsCache)
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
	go t.Listen()

	<-gracefulShutdown

	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// we don't need the context variable here so let's just put underscore
	defer handleTermination(publisher, t, cancel)
}

func handleTermination(publisher *transport.Publisher, t *transport.Consumer, cancel context.CancelFunc) {
	fmt.Println("Start terminating process")

	// close reader
	t.Close()

	// wait 6 seconds until last block will be processed
	time.Sleep(6 * time.Second)

	// close writer
	publisher.Close()

	fmt.Println("Finish")
	cancel()
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
	var provider url.ApiUrlProvider
	if nodeUrl == "" {
		logger.Info("Using trongrid adapter")
		provider = url.NewTrongridUrlProvider()
	} else {
		logger.Info("Using node adapter")
		provider = url.NewNodeUrlProvider(nodeUrl)
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
