package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dllewellyn/reflex/internal/platform/schema"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ConfluentConsumer implements the Consumer interface for Confluent Kafka.
type ConfluentConsumer struct {
	consumer *kafka.Consumer
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(ctx context.Context) (*ConfluentConsumer, error) {
	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	apiKey := os.Getenv("KAFKA_API_KEY")
	apiSecret := os.Getenv("KAFKA_API_SECRET")
	consumerGroupID := os.Getenv("KAFKA_CONSUMER_GROUP_ID")

	if bootstrapServers == "" || apiKey == "" || apiSecret == "" || consumerGroupID == "" {
		return nil, fmt.Errorf("KAFKA environment variables not set")
	}

	securityProtocol := os.Getenv("KAFKA_SECURITY_PROTOCOL")
	if securityProtocol == "" {
		securityProtocol = "SASL_SSL"
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"security.protocol":  securityProtocol,
		"sasl.mechanisms":    "PLAIN",
		"sasl.username":      apiKey,
		"sasl.password":      apiSecret,
		"group.id":           consumerGroupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &ConfluentConsumer{consumer: c}, nil
}

// Consume consumes messages from the Kafka topic and invokes the handler for each message.
// It blocks until the context is cancelled or an error occurs.
func (c *ConfluentConsumer) Consume(ctx context.Context, topic string, handler func(ctx context.Context, msg *schema.InteractionEvent) error) error {
	err := c.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	log.Printf("Kafka consumer subscribed to topic %s, starting to consume messages...", topic)
	run := true
	for run {
		select {
		case <-ctx.Done():
			run = false
		default:
			msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
			if err == nil {
				// Extract the context
				propagator := otel.GetTextMapPropagator()
				carrier := kafkaMessageCarrier{msg: msg}
				ctx := propagator.Extract(context.Background(), carrier)

				// Create a new span
				tr := otel.Tracer("kafka-consumer")
				var span trace.Span
				ctx, span = tr.Start(ctx, "kafka.consume")

				// Unmarshal the message value into the schema
				var event schema.InteractionEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					slog.Warn("Error unmarshaling message", "error", err)
					span.End()
					continue // Skip malformed messages
				}

				if err := handler(ctx, &event); err != nil {
					log.Printf("Error processing message: %v\n", err)
				}
				span.End()
			} else {
				if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
					// Timeout is expected, just continue loop
					continue
				}
				log.Printf("Kafka consumer error: %v\n", err)
			}
		}
	}
	return ctx.Err()
}

// ConsumeBatch consumes messages from the Kafka topic until no messages are received for the specified timeout.
func (c *ConfluentConsumer) ConsumeBatch(ctx context.Context, topic string, handler func(ctx context.Context, msg *schema.InteractionEvent) error, timeout time.Duration) error {
	err := c.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	log.Printf("Kafka consumer subscribed to topic %s (Batch Mode), timeout: %v", topic, timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Proceed
		}

		msg, err := c.consumer.ReadMessage(timeout)
		if err == nil {
			// Extract the context
			propagator := otel.GetTextMapPropagator()
			carrier := kafkaMessageCarrier{msg: msg}
			ctx := propagator.Extract(context.Background(), carrier)

			// Create a new span
			tr := otel.Tracer("kafka-consumer")
			var span trace.Span
			ctx, span = tr.Start(ctx, "kafka.consume")

			// Unmarshal the message value into the schema
			var event schema.InteractionEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling message: %v\n", err)
				span.End()
				continue // Skip malformed messages
			}

			if err := handler(ctx, &event); err != nil {
				log.Printf("Error processing message: %v\n", err)
			}
			span.End()
			continue
		}

		// Check for timeout
		if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
			// Timeout reached, no more messages in this batch
			return nil
		}

		log.Printf("Kafka consumer error: %v\n", err)
	}
}

// Commit commits the offsets of all messages consumed so far.
func (c *ConfluentConsumer) Commit() error {
	// Commit the current assignment offsets.
	// Note: Commit() commits the *last consumed* message's offset + 1.
	// Since we are using manual commit, we should ensure we are committing what we expect.
	// In Confluent Kafka Go, Commit() (sync) or CommitMessage (sync) work.
	// Since we don't have the specific message pointer here readily available without tracking it,
	// Commit() on the consumer instance commits the current partition assignment's stored offsets?
	// Actually, "Commit() commits the current assignment".
	// But we haven't "stored" offsets manually.
	// However, ReadMessage updates the internal state.
	// So calling Commit() should commit the offsets of messages returned by ReadMessage.

	// Wait, we need to be careful. if enable.auto.commit=false, we must explicitly commit.
	// c.consumer.Commit() commits the current position.

	_, err := c.consumer.Commit()
	return err
}

// Close closes the Kafka consumer.
func (c *ConfluentConsumer) Close() error {
	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			log.Printf("Error closing Kafka consumer: %v", err)
			return err
		}
		log.Println("Kafka consumer closed.")
	}
	return nil
}
