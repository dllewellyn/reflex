package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/dllewellyn/reflex/internal/platform/gcs"
	"github.com/dllewellyn/reflex/internal/platform/kafka"
	"github.com/dllewellyn/reflex/internal/platform/schema"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type Config struct {
	Topic string
}

type Service struct {
	consumer  kafka.Consumer
	gcsWriter gcs.BlobWriter
	topic     string
}

func NewService(consumer kafka.Consumer, gcsWriter gcs.BlobWriter, cfg Config) *Service {
	return &Service{
		consumer:  consumer,
		gcsWriter: gcsWriter,
		topic:     cfg.Topic,
	}
}

// RunOnce consumes messages until a timeout occurs (indicating no more immediate messages),
// writes them to GCS, and commits offsets.
func (s *Service) RunOnce(ctx context.Context) error {
	tr := otel.Tracer("loader-service")
	ctx, span := tr.Start(ctx, "RunOnce")
	defer span.End()

	slog.Info("Consuming messages from Kafka...", "topic", s.topic)

	// Buffer to store messages grouped by Conversation ID
	// map[conversationID][]byte (JSON lines)
	sessionBuffers := make(map[string][]schema.InteractionEvent)

	// Consume until 10 seconds of silence
	err := s.consumer.ConsumeBatch(ctx, s.topic, func(ctx context.Context, msg *schema.InteractionEvent) error {
		if msg == nil {
			return nil
		}

		// Group by Conversation ID
		sessionBuffers[msg.ConversationId] = append(sessionBuffers[msg.ConversationId], *msg)
		return nil
	}, 10*time.Second)

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("consumption error: %w", err)
	}

	if len(sessionBuffers) == 0 {
		slog.Info("No messages to process")
		return nil
	}

	slog.Info("Writing batch to GCS", "sessions", len(sessionBuffers))

	// Write each session's buffer to GCS
	for sessionID, events := range sessionBuffers {
		if len(events) == 0 {
			continue
		}

		// Determine time path from the first event (or average? usually grouped by hour so mostly same)
		// Spec: raw/<conversation_id>/YYYY/MM/DD/HH/chunk-<uuid>.jsonl
		t := events[0].Timestamp
		if t.IsZero() {
			t = time.Now()
		}

		key := fmt.Sprintf("raw/%s/%d/%02d/%02d/%02d/chunk-%s.jsonl",
			sessionID, t.Year(), t.Month(), t.Day(), t.Hour(), uuid.New().String())

		// Serialize events to JSONL
		var data []byte
		for _, event := range events {
			b, err := json.Marshal(event)
			if err != nil {
				return fmt.Errorf("marshal error: %w", err)
			}
			data = append(data, b...)
			data = append(data, '\n')
		}

		if err := s.gcsWriter.Write(ctx, key, data); err != nil {
			return fmt.Errorf("gcs write error for session %s: %w", sessionID, err)
		}
	}

	slog.Info("Batch complete")
	if err := s.consumer.Commit(); err != nil {
		slog.Error("Failed to commit offsets", "error", err)
		// We processed and wrote to GCS, so we should probably not fail the job hard if commit fails,
		// but it risks duplicate processing next time.
		// For now return error.
		return fmt.Errorf("failed to commit offsets: %w", err)
	}
	return nil
}
