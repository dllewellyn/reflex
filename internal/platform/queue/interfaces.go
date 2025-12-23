package queue

import (
	"context"
)

// Consumer interface for reading messages from a queue (e.g., Pub/Sub).
type Consumer interface {
	Receive(ctx context.Context, handler func(ctx context.Context, msg []byte) error) error
	Close() error
}

// Producer interface for publishing messages to a queue (e.g., Kafka).
type Producer interface {
	Publish(ctx context.Context, topic string, msg interface{}) error
	Close()
}
