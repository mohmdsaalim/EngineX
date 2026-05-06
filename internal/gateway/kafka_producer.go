package gateway

import (
	"context"
	"fmt"
	"log"

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
	log.Printf("[GATEWAY] Publishing order to Kafka | topic: %s | symbol: %s", constants.TopicOrdersSubmitted, symbol)
	if err := p.producer.Publish(ctx, constants.TopicOrdersSubmitted, symbol, payload); err != nil{
		log.Printf("[GATEWAY] Failed to publish order: %v", err)
		return fmt.Errorf("publish order failed: %w", err)
	}
	log.Printf("[GATEWAY] Order published successfully | topic: %s | symbol: %s", constants.TopicOrdersSubmitted, symbol)
	return nil
}
