package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.opentelemetry.io/otel"
)

type ConfluentProducer struct {
	p *kafka.Producer
}

// NewProducer creates a new ConfluentProducer.
// It returns a pointer to ConfluentProducer which implicitly implements the Producer interface.
func NewProducer(cfg *kafka.ConfigMap) (*ConfluentProducer, error) {
	p, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, err
	}
	return &ConfluentProducer{p: p}, nil
}

// Publish publishes a message to a Kafka topic.
func (k *ConfluentProducer) Publish(ctx context.Context, topic string, key string, msg any) error {
	val, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	deliveryChan := make(chan kafka.Event)
	msgDetails := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          val,
		Headers:        []kafka.Header{},
	}

	// Inject the context
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, kafkaMessageCarrier{msg: msgDetails})

	if key != "" {
		msgDetails.Key = []byte(key)
	}

	err = k.p.Produce(msgDetails, deliveryChan)

	if err != nil {
		return fmt.Errorf("produce error: %w", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	close(deliveryChan)
	return nil
}

func (k *ConfluentProducer) Close() {
	k.p.Flush(15 * 1000)
	k.p.Close()
}
