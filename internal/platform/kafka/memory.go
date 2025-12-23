package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/dllewellyn/reflex/internal/platform/schema"
)

// MemoryProducer is an in-memory implementation of the Producer interface.
type MemoryProducer struct {
	mu       sync.RWMutex
	messages map[string][]Message // topic -> messages
}

// Message represents a message stored in memory.
type Message struct {
	Topic     string
	Key       string
	Value     []byte
	Timestamp time.Time
}

// NewMemoryProducer creates a new in-memory Kafka producer.
func NewMemoryProducer() *MemoryProducer {
	return &MemoryProducer{
		messages: make(map[string][]Message),
	}
}

// Publish publishes a message to an in-memory topic.
func (m *MemoryProducer) Publish(ctx context.Context, topic string, key string, message any) error {
	val, err := json.Marshal(message)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	msg := Message{
		Topic:     topic,
		Key:       key,
		Value:     val,
		Timestamp: time.Now(),
	}

	m.messages[topic] = append(m.messages[topic], msg)
	return nil
}

// GetMessages returns all messages published to a specific topic.
func (m *MemoryProducer) GetMessages(topic string) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := m.messages[topic]
	result := make([]Message, len(messages))
	copy(result, messages)
	return result
}

// Clear clears all messages from all topics.
func (m *MemoryProducer) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make(map[string][]Message)
}

// MemoryConsumer is an in-memory implementation of the Consumer interface.
type MemoryConsumer struct {
	mu       sync.RWMutex
	messages map[string][]schema.InteractionEvent // topic -> events
	offset   map[string]int                       // topic -> current offset
}

// NewMemoryConsumer creates a new in-memory Kafka consumer.
func NewMemoryConsumer() *MemoryConsumer {
	return &MemoryConsumer{
		messages: make(map[string][]schema.InteractionEvent),
		offset:   make(map[string]int),
	}
}

// Seed seeds the consumer with pre-existing messages for testing.
func (m *MemoryConsumer) Seed(topic string, events []schema.InteractionEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages[topic] = append(m.messages[topic], events...)
}

// Consume consumes messages from an in-memory topic and invokes the handler for each message.
func (m *MemoryConsumer) Consume(ctx context.Context, topic string, handler func(ctx context.Context, msg *schema.InteractionEvent) error) error {
	m.mu.Lock()
	messages := m.messages[topic]
	startOffset := m.offset[topic]
	// Make a copy of the relevant portion of the messages slice to avoid race conditions
	msgsCopy := make([]schema.InteractionEvent, len(messages[startOffset:]))
	copy(msgsCopy, messages[startOffset:])
	m.mu.Unlock()

	for i, event := range msgsCopy {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := handler(ctx, &event); err != nil {
				return err
			}

			m.mu.Lock()
			m.offset[topic] = startOffset + i + 1
			m.mu.Unlock()
		}
	}

	// After consuming all available messages, wait for context cancellation
	// This simulates the real Kafka consumer behavior where Consume blocks
	// until the context is cancelled or an error occurs
	<-ctx.Done()
	return ctx.Err()
}

// ConsumeBatch consumes messages from an in-memory topic until no messages are received (all seeded messages processed).
// Since it's in-memory, we can just process all messages and return.
func (m *MemoryConsumer) ConsumeBatch(ctx context.Context, topic string, handler func(ctx context.Context, msg *schema.InteractionEvent) error, timeout time.Duration) error {
	m.mu.Lock()
	messages := m.messages[topic]
	startOffset := m.offset[topic]
	// Make a copy of the relevant portion of the messages slice to avoid race conditions
	msgsCopy := make([]schema.InteractionEvent, len(messages[startOffset:]))
	copy(msgsCopy, messages[startOffset:])
	m.mu.Unlock()

	for i, event := range msgsCopy {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := handler(ctx, &event); err != nil {
				return err
			}

			m.mu.Lock()
			m.offset[topic] = startOffset + i + 1
			m.mu.Unlock()
		}
	}

	// In memory, we assume "Batch" means "Everything currently in the queue".
	// So we return nil immediately after processing.
	return nil
}

// Clear resets the consumer state.
func (m *MemoryConsumer) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make(map[string][]schema.InteractionEvent)
	m.offset = make(map[string]int)
}

// Commit commits the offsets of all messages consumed so far.
func (m *MemoryConsumer) Commit() error {
	return nil
}
