package batch_test

import (
	"context"
	"testing"
	"time"

	"github.com/dllewellyn/reflex/internal/app/batch"
	"github.com/dllewellyn/reflex/internal/platform/gcs"
	"github.com/dllewellyn/reflex/internal/platform/vertex"
)

// MockProducer implements kafka.Producer (implicitly)
type MockProducer struct{}

func (m *MockProducer) Publish(ctx context.Context, topic string, key string, message any) error {
	return nil
}

func TestService_Run(t *testing.T) {
	// Setup dependencies
	gcsClient := gcs.NewMemoryClient()
	vertexClient := vertex.NewMemoryClient()
	producer := &MockProducer{}

	svc := batch.NewService(batch.Config{
		ProjectID:     "test-project",
		Location:      "us-central1",
		StagingBucket: "staging-bucket",
		OutputBucket:  "output-bucket",
		ModelID:       "test-model",
	}, &batch.Prompt{
		Messages: []batch.Message{{Role: "user", Content: "test {{conversation_transcript}}"}},
	}, gcsClient, gcsClient, vertexClient, producer)

	// Seed GCS with some data for "yesterday"
	yesterday := time.Now().AddDate(0, 0, -1)
	ctx := context.Background()

	// path: raw/session-123/YYYY/MM/DD/HH/chunk.jsonl
	key := "raw/session-123/" + yesterday.Format("2006/01/02/15") + "/chunk-1.jsonl"
	if err := gcsClient.Write(ctx, key, []byte(`{"interaction_id":"1"}`)); err != nil {
		t.Fatalf("failed to seed gcs: %v", err)
	}

	// Run service
	if err := svc.Run(ctx, yesterday); err != nil {
		t.Errorf("Run() error = %v", err)
	}
}
