package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goccy/go-json"
	commonModels "github.com/kattana-io/models/pkg/storage"
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
)

const shutdownTimeout = 5

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	/**
	 * Handle run options like close signals
	 */
	runner := runway.Create()
	logger := runner.Logger()
	abiHolder := abi.Create()
	registerCommandLineFlags()
	mode, topic := getRunningMode()
	redis := runner.Redis()

	quotesFile := helper.NewQuotesFile()
	tokenLists := integrations.NewTokensListProvider()
	sunswapLists := integrations.NewSunswapProvider()
	pairsCache := cache.NewPairsCache(redis)

	logger.Info(fmt.Sprintf("Start parser in %s mode", mode))
	publisher := transport.NewPublisher("parser.sys.parsed", os.Getenv("KAFKA"), logger)
	publisherHolders := transport.NewPublisher("holders_blocks", os.Getenv("KAFKA"), logger)
	t := transport.CreateConsumer(topic, logger)
	t.OnBlock(func(Value []byte) bool {
		/**
		 * Decode block
		 */
		block := commonModels.Block{}
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
		api := createAPI(block.Node)
		fiatConverter := converters.CreateConverter(redis, logger, &block, quotesFile.Get())
		p := parser.New(api, tokenLists, pairsCache, fiatConverter, abiHolder, sunswapLists)
		ok := p.Parse(block)
		if ok {
			encodedHolders := p.GetEncodedHolders()
			p.DeleteHolders()
			publisher.PublishBlock(context.Background(), p.GetEncodedBlock())
			publisherHolders.PublishBlock(context.Background(), encodedHolders)
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
	handleTermination(publisher, t)
}

func handleTermination(publisher *transport.Publisher, reader *transport.Consumer) {
	zap.L().Info("Start terminating process")
	reader.Close()
	publisher.Close()
	time.Sleep(shutdownTimeout * time.Second)
	zap.L().Info("Finish")
}

func createAPI(nodeURL string) *tronApi.API {
	var provider url.APIURLProvider
	if nodeURL == "" {
		zap.L().Info("Using trongrid adapter")
		provider = url.NewTrongridURLProvider()
	} else {
		zap.L().Info("Using node adapter")
		provider = url.NewNodeURLProvider(nodeURL)
	}
	api := tronApi.NewAPI(nodeURL, zap.L(), provider)
	return api
}

// getRunningMode - decide should we consume live or history
func getRunningMode() (mode string, topic models.Topics) {
	mode = viper.GetString("mode")

	if mode == "HISTORY" {
		topic = models.TronHistory
	} else {
		topic = models.TronLive
	}
	return mode, topic
}

func registerCommandLineFlags() {
	rootCmd := &cobra.Command{}
	rootCmd.Flags().String("mode", string(models.LIVE), "Please provide mode: --mode LIVE or --mode HISTORY")

	err := viper.BindPFlag("mode", rootCmd.Flags().Lookup("mode"))
	if err != nil {
		zap.L().Fatal("registerCommandLineFlags: BindPFlag", zap.Error(err))
	}

	err = rootCmd.Execute()
	if err != nil {
		zap.L().Fatal("registerCommandLineFlags: Execute", zap.Error(err))
	}
}
