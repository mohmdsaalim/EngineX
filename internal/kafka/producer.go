package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Balancer:    &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:       true,
		AllowAutoTopicCreation: true,
		// Internal retry via async batch settings
		BatchSize:  1,
		BatchTimeout: 100,
		ErrorLogger:  kafka.LoggerFunc(func(s string, i ...interface{}) {}),
	}
	return &Producer{writer: writer}
}

func (p *Producer) Publish(ctx context.Context, topic, key string, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}