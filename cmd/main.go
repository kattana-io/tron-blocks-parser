package main

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/cache"
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
	registerCommandLineFlags(logger)
	mode, topic := getRunningMode()
	redis := runner.Redis()
	runner.Run()

	api := createApi(logger)
	tokenLists := integrations.NewTokensListProvider(logger)
	pairsCache := cache.NewPairsCache(redis, logger)

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
		 * Process block
		 */
		//fiatConverter := parser.NewFiatConverter(redis, logger, block.Number)
		p := parser.New(api, logger, tokenLists, pairsCache)
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

func createApi(logger *zap.Logger) *tronApi.Api {
	nodeUrl := os.Getenv("SOLIDITY_FULL_NODE_URL")
	if nodeUrl == "" {
		logger.Fatal("Empty SOLIDITY_FULL_NODE_URL, can't continue")
	}
	api := tronApi.NewApi(nodeUrl, logger)
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
