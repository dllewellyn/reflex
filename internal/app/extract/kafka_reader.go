package extract

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dllewellyn/reflex/internal/platform/schema"
)

type KafkaResultReader struct {
	config Config
}

func NewKafkaResultReader(cfg Config) *KafkaResultReader {
	return &KafkaResultReader{
		config: cfg,
	}
}

type LogWrapper struct {
	Time  string                  `json:"time"`
	Level string                  `json:"level"`
	Msg   string                  `json:"msg"`
	Event schema.BatchResultEvent `json:"event"`
}

func (k *KafkaResultReader) ReadResults(ctx context.Context) (<-chan BatchResult, <-chan error, func()) {
	out := make(chan BatchResult)
	errCh := make(chan error, 1)

	// We need to keep track of the consumer to close it later
	var consumer *kafka.Consumer

	// Cleanup function
	teardown := func() {
		if consumer != nil {
			consumer.Close()
		}
	}

	go func() {
		defer close(out)
		defer close(errCh)

		// Initialize Consumer
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  k.config.KafkaBootstrapServers,
			"security.protocol":  "SASL_SSL",
			"sasl.mechanisms":    "PLAIN",
			"sasl.username":      k.config.KafkaAPIKey,
			"sasl.password":      k.config.KafkaAPISecret,
			"group.id":           "extract-injections-consumer",
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": false,
		})
		if err != nil {
			errCh <- fmt.Errorf("failed to create kafka consumer: %w", err)
			return
		}
		consumer = c // Assign to outer variable for cleanup

		if err := c.SubscribeTopics([]string{k.config.KafkaTopic}, nil); err != nil {
			errCh <- fmt.Errorf("failed to subscribe to topic %s: %w", k.config.KafkaTopic, err)
			return
		}

		slog.Info("Started Kafka consumer", "topic", k.config.KafkaTopic, "group", "extract-injections-consumer")

		lastMsgTime := time.Now()
		timeoutDuration := time.Duration(k.config.IdleTimeoutSeconds) * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
						if time.Since(lastMsgTime) > timeoutDuration {
							slog.Info("Idle timeout reached, shutting down consumer", "timeout_seconds", k.config.IdleTimeoutSeconds)
							return
						}
						continue
					}
					slog.Error("Kafka read error", "error", err)
					// Verify if we should stop or continue on error.
					// For now log and continue
					continue
				}
				lastMsgTime = time.Now()

				var event schema.BatchResultEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					slog.Warn("Failed to unmarshal event", "error", err, "raw", string(msg.Value))
					continue
				}

				// The record field is now strongly typed in the generated code.
				// We can access it directly.

				slog.Info("Received event", "event_id", event.EventId)

				// Validation logic adjusted for the new schema structure
				if event.Record.Request == nil || len(event.Record.Request.Contents) == 0 {
					slog.Warn("Validation failed: Request has no contents", "event_id", event.EventId, "source", event.Source)
				} else if len(event.Record.Request.Contents[0].Parts) == 0 {
					slog.Warn("Validation failed: Request content has no parts", "event_id", event.EventId, "source", event.Source)
				}

				if event.Record.Response == nil || len(event.Record.Response.Candidates) == 0 {
					slog.Warn("Validation failed: Response has no candidates", "event_id", event.EventId, "source", event.Source)
				} else if len(event.Record.Response.Candidates[0].Content.Parts) == 0 {
					slog.Warn("Validation failed: Response candidate has no content parts", "event_id", event.EventId, "source", event.Source)
				}

				// Safely access fields given pointers in generated code
				var candidates []Candidate
				if event.Record.Response != nil {
					for _, c := range event.Record.Response.Candidates {
						// c.Content is a value of type CandidateContent
						// We need to map it to our domain Content (which has Parts []Part)
						// Our domain Part has Text string.

						var parts []Part
						for _, p := range c.Content.Parts {
							// p is ContentPartsPart, p.Text is string
							parts = append(parts, Part{Text: p.Text})
						}
						// domain.Content
						content := Content{Parts: parts}
						candidates = append(candidates, Candidate{Content: content})
					}
				}

				var contents []Content
				if event.Record.Request != nil {
					for _, c := range event.Record.Request.Contents {
						var parts []Part
						for _, p := range c.Parts {
							if p.Text != nil {
								// p.Text in generated Request/Content/Part is *string
								parts = append(parts, Part{Text: *p.Text})
							}
						}
						contents = append(contents, Content{Parts: parts})
					}
				}

				result := BatchResult{
					EventID:  event.EventId,
					Response: Response{Candidates: candidates},
					Request:  Request{Contents: contents},
					Commit: func() {
						if _, err := c.CommitMessage(msg); err != nil {
							slog.Error("Failed to commit message", "error", err, "event_id", event.EventId)
						}
					},
				}

				select {
				case out <- result:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, errCh, teardown
}
