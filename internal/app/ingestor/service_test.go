package ingestor

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dllewellyn/reflex/internal/platform/pinecone"
	"github.com/dllewellyn/reflex/internal/platform/schema"
)

type MockProducer struct {
	PublishFunc func(ctx context.Context, topic string, key string, msg interface{}) error
}

func (m *MockProducer) Publish(ctx context.Context, topic string, key string, msg interface{}) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, key, msg)
	}
	return nil
}

func (m *MockProducer) Close() {}

type MockVectorStore struct {
	QueryInputFunc func(ctx context.Context, text string, topK int) ([]*pinecone.Match, error)
}

func (m *MockVectorStore) UpsertBatch(ctx context.Context, vectors []*pinecone.Vector) error {
	return nil
}
func (m *MockVectorStore) UpsertInputs(ctx context.Context, inputs []*pinecone.InputRecord) error {
	return nil
}
func (m *MockVectorStore) QueryInput(ctx context.Context, text string, topK int) ([]*pinecone.Match, error) {
	if m.QueryInputFunc != nil {
		return m.QueryInputFunc(ctx, text, topK)
	}
	return nil, nil // Return empty matches by default
}
func (m *MockVectorStore) Fetch(ctx context.Context, ids []string) (map[string]*pinecone.Vector, error) {
	return nil, nil
}
func (m *MockVectorStore) DescribeIndexStats(ctx context.Context) (*pinecone.IndexStats, error) {
	return &pinecone.IndexStats{}, nil
}
func (m *MockVectorStore) DeleteAll(ctx context.Context) error {
	return nil
}

func TestAnalyzeInteraction(t *testing.T) {
	publishedMain := false
	mockProducer := &MockProducer{
		PublishFunc: func(ctx context.Context, topic string, key string, msg interface{}) error {
			if topic == "test-topic" {
				publishedMain = true
			} else {
				t.Errorf("unexpected topic %s", topic)
			}
			event, ok := msg.(schema.InteractionEvent)
			if !ok {
				t.Errorf("expected schema.InteractionEvent, got %T", msg)
			}
			if event.InteractionId != "123" {
				t.Errorf("expected id 123, got %s", event.InteractionId)
			}
			if event.Content != "analyze me" {
				t.Errorf("expected content 'analyze me', got %s", event.Content)
			}
			return nil
		},
	}

	svc := NewService(mockProducer, &MockVectorStore{}, Config{TopicName: "test-topic", Port: "8080"})

	interaction := map[string]interface{}{
		"interaction_id":  "123",
		"conversation_id": "456",
		"prompt":          "analyze me",
		"user": map[string]string{
			"user_id": "user-1",
		},
	}
	body, _ := json.Marshal(interaction)
	req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
	w := httptest.NewRecorder()

	svc.AnalyzeInteraction(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Result().StatusCode)
	}

	if !publishedMain {
		t.Error("expected message to be published to main topic")
	}
}

func TestRun(t *testing.T) {
	svc := NewService(&MockProducer{}, &MockVectorStore{}, Config{TopicName: "test-topic", Port: "0"}) // 0 for random port

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)

	go func() {
		errChan <- svc.Run(ctx)
	}()

	// Let it start
	// In a real scenario, we might want to wait for the port to be open,
	// but here we just want to ensure it doesn't immediately fail.
	// We'll just wait a tiny bit.
	cancel()

	err := <-errChan
	if err != nil {
		t.Errorf("expected nil error on graceful shutdown, got %v", err)
	}
}
