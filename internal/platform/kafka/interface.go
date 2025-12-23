package kafka

import (
	"context"
	"time"

	"github.com/dllewellyn/reflex/internal/platform/schema"
)

// Producer defines the interface for publishing messages to Kafka.
type Producer interface {
	// Publish publishes a message to a Kafka topic.
	// The message can be any type that the implementation can serialize (e.g., struct, []byte).
	Publish(ctx context.Context, topic string, key string, message any) error
}

// Consumer defines the interface for consuming messages from Kafka.
type Consumer interface {
	// Consume consumes messages from a Kafka topic and invokes the handler for each message.
	// It blocks until the context is cancelled or an error occurs.
	Consume(ctx context.Context, topic string, handler func(ctx context.Context, msg *schema.InteractionEvent) error) error

	// ConsumeBatch consumes messages from a Kafka topic until no messages are received for the specified timeout.
	ConsumeBatch(ctx context.Context, topic string, handler func(ctx context.Context, msg *schema.InteractionEvent) error, timeout time.Duration) error
	// Commit commits the offsets of all messages consumed so far.
	Commit() error
}
