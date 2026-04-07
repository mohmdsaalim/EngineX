package gateway

import (
	"context"
	"fmt"

	"github.com/mohmdsaalim/EngineX/internal/constants"
	"github.com/mohmdsaalim/EngineX/internal/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer( producer *kafka.Producer ) *KafkaProducer {
	return &KafkaProducer{producer: producer}
}

func (p *KafkaProducer) PublishOrder(ctx context.Context, symbol string, payload []byte) error{
	if err := p.producer.Publish(ctx, constants.TopicOrdersSubmitted, symbol, payload); err != nil{
		return fmt.Errorf("publish order failed: %w", err)
	}
	return nil
}
