package transport

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader  *kafka.Reader
	log     *zap.Logger
	blockFn func(Value []byte) bool
	failFn  func(error error)
}

func (t *Consumer) Listen() {
	for {
		m, err := t.reader.ReadMessage(context.Background())
		if err != nil {
			t.failFn(err)
			break
		}
		t.blockFn(m.Value)
	}

	if err := t.reader.Close(); err != nil {
		t.failFn(err)
	}
}

func (t *Consumer) OnFail(f func(error error)) {
	t.failFn = f
}

func (t *Consumer) OnBlock(f func(Value []byte) bool) {
	t.blockFn = f
}

func (t *Consumer) Close() {
	err := t.reader.Close()

	if err != nil {
		t.log.Error(fmt.Sprintf("Error during reader closing: %s", err.Error()))
	}
}

func CreateConsumer(topic models.Topics, log *zap.Logger) *Consumer {
	brokers := strings.Split(os.Getenv("KAFKA"), ",")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    string(topic),
		GroupID:  "parsers",
		MinBytes: 1e3,  // 1KB
		MaxBytes: 50e6, // 50MB
		MaxWait:  1 * time.Second,
	})

	log.Info("Connected to topic: " + string(topic))

	return &Consumer{
		reader: reader,
		log:    log,
	}
}
