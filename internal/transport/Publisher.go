package transport

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
	models "github.com/kattana-io/models/pkg/storage"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Publisher struct {
	log     *zap.Logger
	w       *kafka.Writer
	address []string
}

func NewPublisher(topic string, address []string, log *zap.Logger) *Publisher {
	w := &kafka.Writer{
		Addr:     kafka.TCP(address...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}

	return &Publisher{
		log:     log,
		w:       w,
		address: address,
	}
}

func (p *Publisher) PublishBlock(ctx context.Context, block []byte) {
	if err := p.w.WriteMessages(ctx, kafka.Message{Value: block}); err != nil {
		p.log.Fatal(fmt.Sprintf("failed to write messages: %s", err.Error()))
	}
}

func (p *Publisher) Close() {
	if err := p.w.Close(); err != nil {
		p.log.Fatal(fmt.Sprintf("failed to close writer: %s", err.Error()))
	}
}

// PublishFailedBlock Create a temporary failed publisher and return block to sender
func (p *Publisher) PublishFailedBlock(ctx context.Context, block models.Block) bool {
	failedBlocksWriter := &kafka.Writer{
		Addr:     kafka.TCP(p.address...),
		Topic:    "failed_blocks",
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}
	Value, err := json.Marshal(block)
	if err != nil {
		p.log.Error(err.Error())
		return false
	}
	if err := failedBlocksWriter.WriteMessages(ctx, kafka.Message{Value: Value}); err != nil {
		p.log.Fatal(fmt.Sprintf("failed to write messages: %s", err.Error()))
		return false
	}
	return true
}
