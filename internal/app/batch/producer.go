package batch

import (
	"context"

	"github.com/dllewellyn/reflex/internal/platform/schema"
)

// EventPublisher defines the interface for publishing events.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, msg any) error
}

// BatchEventProducer produces BatchResultEvents to Kafka.
type BatchEventProducer struct {
	publisher EventPublisher
	topic     string
}

// NewBatchEventProducer creates a new BatchEventProducer.
func NewBatchEventProducer(publisher EventPublisher, topic string) *BatchEventProducer {
	return &BatchEventProducer{
		publisher: publisher,
		topic:     topic,
	}
}

// Produce publishes a batch result event.
func (p *BatchEventProducer) Produce(ctx context.Context, event schema.BatchResultEvent) error {
	// Use eventId as key for partitioning
	key := event.EventId
	return p.publisher.Publish(ctx, p.topic, key, event)
}
