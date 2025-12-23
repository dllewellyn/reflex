package features

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dllewellyn/reflex/internal/platform/kafka"
	"github.com/dllewellyn/reflex/internal/platform/schema"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// getTestTopic returns the configured Kafka topic or a default for testing.
func getTestTopic() string {
	t := os.Getenv("KAFKA_TOPIC")

	isRealTest := os.Getenv("TEST_USE_REAL_KAFKA") == "true"

	if isRealTest && t == "" {
		// if we're using real kafka, this must be set so log a fatal error
		log.Fatal("KAFKA_TOPIC must be set for real kafka tests")
	}

	if t == "" {
		return "raw-interactions"
	}
	return t
}

type TestInfrastructure interface {
	GetProducer() kafka.Producer
	// StartConsumer prepares the consumer to listen for messages.
	// Must be called before the action that produces messages.
	StartConsumer(t *testing.T, topic string)
	// VerifyMessagePublished checks that a message with the expected ID was published.
	VerifyMessagePublished(t *testing.T, topic string, expectedID string)
	// VerifyNoMessagePublished checks that no messages were published (best effort).
	VerifyNoMessagePublished(t *testing.T, topic string)
	Close()
}

// setupTest returns the appropriate TestInfrastructure based on environment.
func setupTest(t *testing.T) TestInfrastructure {
	if os.Getenv("TEST_USE_REAL_KAFKA") == "true" {
		return NewRealKafkaInfrastructure(t)
	}
	return NewMemoryInfrastructure()
}

// MemoryInfrastructure for fast, isolated tests.
type MemoryInfrastructure struct {
	producer *kafka.MemoryProducer
}

func NewMemoryInfrastructure() *MemoryInfrastructure {
	return &MemoryInfrastructure{
		producer: kafka.NewMemoryProducer(),
	}
}

func (m *MemoryInfrastructure) GetProducer() kafka.Producer {
	return m.producer
}

func (m *MemoryInfrastructure) StartConsumer(t *testing.T, topic string) {
	// No-op for memory, we just check the internal state later
}

func (m *MemoryInfrastructure) VerifyMessagePublished(t *testing.T, topic string, expectedID string) {
	messages := m.producer.GetMessages(topic)
	found := false
	for _, msg := range messages {
		var interaction schema.Interaction
		if err := json.Unmarshal(msg.Value, &interaction); err == nil {
			if interaction.InteractionID == expectedID {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("expected message with interaction_id %s not found in topic %s", expectedID, topic)
	}
}

func (m *MemoryInfrastructure) VerifyNoMessagePublished(t *testing.T, topic string) {
	messages := m.producer.GetMessages(topic)
	if len(messages) > 0 {
		t.Errorf("expected 0 messages in topic %s, got %d", topic, len(messages))
	}
}

func (m *MemoryInfrastructure) Close() {}

// RealKafkaInfrastructure for integration tests.
// RealKafkaInfrastructure for integration tests.
// RealKafkaInfrastructure for integration tests.
type RealKafkaInfrastructure struct {
	producer      *kafka.ConfluentProducer
	consumer      *ckafka.Consumer
	receivedMsgs  chan *ckafka.Message
	cancelConsume context.CancelFunc
}

func NewRealKafkaInfrastructure(t *testing.T) *RealKafkaInfrastructure {
	// Load .env from project root (assuming tests run from features/ dir)
	_ = godotenv.Load("../.env")

	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if bootstrapServers == "" {
		t.Fatal("KAFKA_BOOTSTRAP_SERVERS must be set for real kafka tests (check your .env file)")
	}

	// Config for Producer
	pCfg := &ckafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	}

	// Only add SASL if we have credentials (likely not for local container)
	if os.Getenv("KAFKA_API_KEY") != "" {
		_ = pCfg.SetKey("sasl.username", os.Getenv("KAFKA_API_KEY"))
		_ = pCfg.SetKey("sasl.password", os.Getenv("KAFKA_API_SECRET"))
		_ = pCfg.SetKey("security.protocol", "SASL_SSL")
		_ = pCfg.SetKey("sasl.mechanisms", "PLAIN")
	} else {
		// For local container/plaintext
		_ = pCfg.SetKey("security.protocol", "PLAINTEXT")
	}

	p, err := kafka.NewProducer(pCfg)
	if err != nil {
		t.Fatalf("Failed to create real producer: %v", err)
	}

	return &RealKafkaInfrastructure{
		producer:     p,
		receivedMsgs: make(chan *ckafka.Message, 100),
	}
}

func (r *RealKafkaInfrastructure) GetProducer() kafka.Producer {
	return r.producer
}

func (r *RealKafkaInfrastructure) StartConsumer(t *testing.T, topic string) {
	// Create a unique consumer group for this test run to ensure we see new messages
	groupID := fmt.Sprintf("test-group-%s", uuid.New().String())

	brokers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")

	cCfg := &ckafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest", // For tests, we want everything sent during the test
	}

	if os.Getenv("KAFKA_API_KEY") != "" {
		_ = cCfg.SetKey("sasl.username", os.Getenv("KAFKA_API_KEY"))
		_ = cCfg.SetKey("sasl.password", os.Getenv("KAFKA_API_SECRET"))
		_ = cCfg.SetKey("security.protocol", "SASL_SSL")
		_ = cCfg.SetKey("sasl.mechanisms", "PLAIN")
	} else {
		_ = cCfg.SetKey("security.protocol", "PLAINTEXT")
	}

	c, err := ckafka.NewConsumer(cCfg)
	if err != nil {
		t.Fatalf("Failed to create real consumer: %v", err)
	}
	r.consumer = c

	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Start consuming in background
	ctx, cancel := context.WithCancel(context.Background())
	r.cancelConsume = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := c.ReadMessage(100 * time.Millisecond)
				if err == nil {
					r.receivedMsgs <- msg
				}
			}
		}
	}()
}

func (r *RealKafkaInfrastructure) VerifyMessagePublished(t *testing.T, topic string, expectedID string) {
	timeout := time.After(10 * time.Second) // generous timeout for real network
	found := false

	for !found {
		select {
		case msg := <-r.receivedMsgs:
			var interaction schema.Interaction
			if err := json.Unmarshal(msg.Value, &interaction); err == nil {
				if interaction.InteractionID == expectedID {
					found = true
				}
			}
		case <-timeout:
			t.Errorf("Timed out waiting for message %s on real Kafka topic %s", expectedID, topic)
			return
		}
	}
}

func (r *RealKafkaInfrastructure) VerifyNoMessagePublished(t *testing.T, topic string) {
	// Wait a bit to ensure nothing arrives
	select {
	case msg := <-r.receivedMsgs:
		t.Errorf("Expected no messages, but got one: %s", string(msg.Value))
	case <-time.After(2 * time.Second):
		// Success
	}
}

func (r *RealKafkaInfrastructure) Close() {
	if r.cancelConsume != nil {
		r.cancelConsume()
	}
	if r.consumer != nil {
		_ = r.consumer.Close()
	}
	if r.producer != nil {
		r.producer.Close()
	}
}
