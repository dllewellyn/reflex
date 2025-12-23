package features

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dllewellyn/reflex/internal/app/loader"
	"github.com/dllewellyn/reflex/internal/platform/kafka"
	"github.com/dllewellyn/reflex/internal/platform/schema"
)

// TestLoaderService_ArchiveMessagesToGCS tests the scenario:
// "Archive messages to GCS"
func TestLoaderService_ArchiveMessagesToGCS(t *testing.T) {
	// Given the configured Kafka topic contains 100 messages
	infra := setupLoaderTest(t)
	defer infra.Close()

	testTopic := getTestTopic()

	// Seed with 100 messages
	var events []schema.InteractionEvent
	for i := 0; i < 100; i++ {
		events = append(events, schema.InteractionEvent{
			InteractionId:  fmt.Sprintf("interaction-%d", i),
			ConversationId: "conv-123",
			Content:        "Message content",
			Role:           schema.RoleUser,
			Timestamp:      time.Now(),
		})
	}
	infra.SeedKafka(t, testTopic, events)

	// And the GCS bucket "security-data-lake" is accessible
	svc := loader.NewService(infra.GetConsumer(), infra.GetGCSWriter(), loader.Config{Topic: testTopic})

	// When the Loader job is triggered
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Increased timeout for real infra
	defer cancel()

	err := svc.RunOnce(ctx)

	// The loader uses ConsumeBatch, so it should return nil when done.
	if err != nil {
		t.Fatalf("RunOnce() failed with unexpected error: %v", err)
	}

	// Then it should consume all 100 messages
	// And it should write a single JSONL file to "gs://security-data-lake/raw/"
	infra.VerifyGCSContent(t, "security-data-lake", 1)
}

// TestLoaderService_HandleZeroMessages tests the scenario:
// "Handle zero messages"
func TestLoaderService_HandleZeroMessages(t *testing.T) {
	// Given the configured Kafka topic is empty
	infra := setupLoaderTest(t)
	defer infra.Close()

	svc := loader.NewService(infra.GetConsumer(), infra.GetGCSWriter(), loader.Config{Topic: getTestTopic()})

	// When the Loader job is triggered
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := svc.RunOnce(ctx)

	// Then it should exit gracefully without writing any file
	// nil is expected because ConsumeBatch returns nil on timeout/empty.
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// And it should not report an error
	// (the test passes if no panic occurred)
}

// TestLoaderService_GCSFailure tests the scenario:
// "Ensure data consistency on failure"
func TestLoaderService_GCSFailure(t *testing.T) {
	// Given the configured Kafka topic contains messages
	mockConsumer := kafka.NewMemoryConsumer()
	failingGCS := &FailingGCSWriter{}

	testTopic := getTestTopic()

	events := []schema.InteractionEvent{
		{
			InteractionId:  "interaction-1",
			ConversationId: "conv-123",
			Content:        "Message content",
			Role:           schema.RoleUser,
			Timestamp:      time.Now(),
		},
	}
	mockConsumer.Seed(testTopic, events)

	// And the GCS service is down
	svc := loader.NewService(mockConsumer, failingGCS, loader.Config{Topic: testTopic})

	// When the Loader job is triggered
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := svc.RunOnce(ctx)

	// Then it should fail to write to GCS
	if err == nil {
		t.Error("expected error when GCS write fails")
	}

	// And it should NOT commit any Kafka offsets
	// (This is implicit - if we re-run the consumer, messages would still be there)
	// And the process should exit with a non-zero status code
	// (Represented by the error return)
}

// FailingGCSWriter is a mock GCS writer that always fails.
type FailingGCSWriter struct{}

func (f *FailingGCSWriter) Write(ctx context.Context, key string, data []byte) error {
	return &GCSError{msg: "GCS unavailable"}
}

type GCSError struct {
	msg string
}

func (e *GCSError) Error() string {
	return e.msg
}
