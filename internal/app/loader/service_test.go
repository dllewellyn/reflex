package loader_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dllewellyn/reflex/internal/app/loader"
	"github.com/dllewellyn/reflex/internal/platform/schema"
)

// --- Mocks ---

type MockConsumer struct {
	Messages []*schema.InteractionEvent
	mu       sync.Mutex
}

func (m *MockConsumer) Consume(ctx context.Context, topic string, handler func(context.Context, *schema.InteractionEvent) error) error {
	m.mu.Lock()
	msgs := m.Messages
	m.mu.Unlock()

	// Simulate sending messages
	for _, msg := range msgs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := handler(ctx, msg); err != nil {
			return err
		}
		// Small sleep to simulate processing time, optional
		time.Sleep(10 * time.Millisecond)
	}

	// Block until context is canceled (simulating waiting for more messages)
	<-ctx.Done()
	return ctx.Err()
}

func (m *MockConsumer) ConsumeBatch(ctx context.Context, topic string, handler func(context.Context, *schema.InteractionEvent) error, timeout time.Duration) error {
	m.mu.Lock()
	msgs := m.Messages
	m.mu.Unlock()

	for _, msg := range msgs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := handler(ctx, msg); err != nil {
			return err
		}
	}
	// Batch done
	return nil
}

func (m *MockConsumer) Commit() error {
	return nil
}

type MockGCSWriter struct {
	Writes map[string][]byte
	mu     sync.Mutex
}

func (m *MockGCSWriter) Write(ctx context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Writes == nil {
		m.Writes = make(map[string][]byte)
	}
	m.Writes[key] = data
	return nil
}

// --- Tests ---

func TestService_RunOnce(t *testing.T) {
	// Setup Data
	sessionID := "test-session-123"
	now := time.Now()
	events := []*schema.InteractionEvent{
		{
			InteractionId:  "int-1",
			ConversationId: sessionID,
			Timestamp:      now,
			Role:           "user",
			Content:        "Hello",
		},
		{
			InteractionId:  "int-2",
			ConversationId: sessionID,
			Timestamp:      now.Add(1 * time.Second),
			Role:           "model",
			Content:        "Hi there",
		},
	}

	mockConsumer := &MockConsumer{Messages: events}
	mockWriter := &MockGCSWriter{}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "test-topic"
	}
	svc := loader.NewService(mockConsumer, mockWriter, loader.Config{Topic: topic})

	// We need to override the silence timer duration in the service for testing,
	// but the service hardcodes it to 10s.
	// We can't easily change it without refactoring.
	// For this test, we can use a long timeout on the test itself, or rely on the mock
	// returning quickly if we didn't implement the "Block until Done" behavior.
	// BUT, the service logic RELIES on the silence timer to cancel the context.
	//
	// To make the test run fast, we should probably refactor the service to accept a config
	// or timeout duration.
	// However, I'm constrained to "Test T004" and minimal changes.
	// I'll try to execute RunOnce. Since the hardcoded timer is 10s, the test will take at least 10s.
	// That's acceptable for a "RunOnce" semantics test.

	ctx := context.Background()

	err := svc.RunOnce(ctx)

	if err != nil {
		t.Fatalf("RunOnce failed: %v", err)
	}

	// Verify Writes
	if len(mockWriter.Writes) == 0 {
		t.Fatal("Expected writes to GCS, got none")
	}

	// Check Key Format
	var foundKey string
	for k := range mockWriter.Writes {
		if strings.Contains(k, sessionID) {
			foundKey = k
			break
		}
	}
	if foundKey == "" {
		t.Errorf("Expected GCS key containing session ID %s, got keys: %v", sessionID, mockWriter.Writes)
	}

	// Check Content
	data := mockWriter.Writes[foundKey]
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 JSONL lines, got %d", len(lines))
	}

	var loadedEvent schema.InteractionEvent
	if err := json.Unmarshal([]byte(lines[0]), &loadedEvent); err != nil {
		t.Fatalf("Failed to unmarshal first line: %v", err)
	}
	if loadedEvent.InteractionId != "int-1" {
		t.Errorf("Expected first event ID int-1, got %s", loadedEvent.InteractionId)
	}
}
