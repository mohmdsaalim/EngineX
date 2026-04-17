package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(broker, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{broker},
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: 0, // manual commit only
		}),
	}
}

// ReadMessage fetches one message — does NOT commit offset.
func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("fetch message: %w", err)
	}
	return msg, nil
}

// CommitMessage commits offset only after successful processing.
// Critical — prevents data loss on restart.
func (c *Consumer) CommitMessage(ctx context.Context, msg kafka.Message) error {
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return fmt.Errorf("commit message: %w", err)
	}
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}