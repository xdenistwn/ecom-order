package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"order/models"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	Writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) Close() error {
	return p.Writer.Close()
}

func (p *KafkaProducer) PublishOrderCreated(ctx context.Context, event interface{}) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", event.(models.OrderCreatedEvent).OrderID)),
		Value: value,
	}

	return p.Writer.WriteMessages(ctx, msg)
}
