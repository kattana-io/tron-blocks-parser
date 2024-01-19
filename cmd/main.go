package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/goccy/go-json"
	"github.com/kattana-io/mesh/pkg/mesh"
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
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const shutdownTimeout = 5

func main() {
	appCtx, cancel := context.WithCancel(context.Background())
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

	errSig := make(chan error)
	connector, err := mesh.Auto(appCtx, errSig)
	if err != nil {
		zap.L().Fatal("Could not connect to the queue", zap.Error(err))
	}

	brokerAddr := strings.Split(os.Getenv("KAFKA"), ",")
	publisher := transport.NewPublisher("parser.sys.parsed", brokerAddr, logger)
	publisherHolders := transport.NewPublisher("holders_blocks", brokerAddr, logger)

	publisherChan := make(chan []byte, 1)
	go connector.ReadIntoChannel(kafka.ReaderConfig{
		Topic:    string(topic),
		GroupID:  "parsers",
		MinBytes: 1e3,  // 1KB
		MaxBytes: 50e6, // 10MB
		MaxWait:  1 * time.Second,
	}, publisherChan)

	go func() {
		for {
			select {
			case msg := <-publisherChan:
				/**
				 * Decode block
				 */
				block := commonModels.Block{}
				err := json.Unmarshal(msg, &block)
				if err != nil {
					logger.Error(err.Error())
					continue
				}

				/**
				 * Check for valid block number
				 * Why we can have 0 here? Invalid value in JSON
				 */
				if block.Number.Int64() == 0 {
					logger.Info("Received null block number, skipping")
					continue
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
					publisher.PublishBlock(appCtx, p.GetEncodedBlock())
					publisherHolders.PublishBlock(appCtx, encodedHolders)
					continue
				} else {
					publisher.PublishFailedBlock(appCtx, block)
				}
			case <-appCtx.Done():
				zap.L().Info("gracefully closing app")
				break
			case err := <-errSig:
				zap.L().Error("Exiting because can't read from queue", zap.Error(err))
				zap.L().Error("Read errors: ", zap.Errors("errors", connector.Errors()))
				gracefulShutdown <- os.Interrupt
			}
		}
	}()

	<-gracefulShutdown
	cancel()
	handleTermination(publisher)
}

func handleTermination(publisher *transport.Publisher) {
	zap.L().Info("Start terminating process")
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
