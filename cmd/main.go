package main

import (
	"fmt"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"github.com/kattana-io/tron-blocks-parser/internal/runway"
	"github.com/kattana-io/tron-blocks-parser/internal/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	/**
	 * Handle run options like close signals
	 */
	runner := runway.Create()
	logger := runner.Logger()
	registerCommandLineFlags(logger)
	mode := viper.GetString("mode")
	//redis := runner.Redis()
	runner.Run()

	var topic models.Topics
	if mode == "HISTORY" {
		topic = models.TRON_HISTORY
	} else {
		topic = models.TRON_LIVE
	}

	logger.Info(fmt.Sprintf("Start parser in %s mode", mode))
	//publisher := transport.NewPublisher("parser.sys.parsed", os.Getenv("KAFKA"), logger)
	t := transport.CreateConsumer(topic, logger)
	t.OnBlock(func(Value []byte) bool {

		// @todo implement onBlock handler
		// use publisher inside
		// redis to cache some inter result
		return true
	})

	t.OnFail(func(err error) {
		logger.Fatal("failed to close reader: " + err.Error())
	})
	t.Listen()
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
