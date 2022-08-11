package transport

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"os"
	"time"
)

type Consumer struct {
	reader    *kafka.Reader
	log       *zap.Logger
	blockFn   func(Value []byte) bool
	failFn    func(error error)
	publisher *Publisher
	redis     *redis.Client
}

func (t *Consumer) Listen() {
	for {
		m, err := t.reader.ReadMessage(context.Background())
		if err != nil {
			t.log.Error(err.Error())
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

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{os.Getenv("KAFKA")},
		Topic:          string(topic),
		GroupID:        "parsers",
		MinBytes:       1e3,         // 1KB
		MaxBytes:       50e6,        // 50MB
		CommitInterval: time.Second, // flushes commits to Kafka every second
	})

	log.Info("Connected to topic: " + string(topic))

	return &Consumer{
		reader: reader,
		log:    log,
	}
}
