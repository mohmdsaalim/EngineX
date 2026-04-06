package gateway

import (
	"context"
	"fmt"

	"github.com/mohmdsaalim/EngineX/internal/constants"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}
// NewKafkaProducer creates a Kafka writer for orders.submitted topic.
func NewKafkaProducer(broker string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr: kafka.TCP(broker),
		Topic: constants.TopicOrdersSubmitted,
		Balancer: &kafka.Hash{}, // partition by key
		RequiredAcks: kafka.RequireOne,
		Async: false,
	}
	return &KafkaProducer{writer: writer}
}

// publishorder sneds order to orders.submtd topic
// key = symbol - ensure a;; INFY orders go to same partition
func (p *KafkaProducer) PublishOrder(ctx context.Context, symbol string, payload []byte) error{
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key: []byte(symbol),
		Value: payload,
	})
	if err != nil{
		return fmt.Errorf("publish orders: %w", err)
	}
	return nil
}

// Close shuts down the kafka wrter cleanly
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

