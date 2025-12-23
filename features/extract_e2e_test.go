package features

import (
	"context"
	"os"
	"testing"

	"github.com/dllewellyn/reflex/internal/app/extract"
	"github.com/dllewellyn/reflex/internal/platform/genai"
	"github.com/dllewellyn/reflex/internal/platform/pinecone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPinecone for E2E
type MockPinecone struct {
	mock.Mock
}

func (m *MockPinecone) UpsertBatch(ctx context.Context, vectors []*pinecone.Vector) error {
	args := m.Called(ctx, vectors)
	return args.Error(0)
}

func (m *MockPinecone) UpsertInputs(ctx context.Context, inputs []*pinecone.InputRecord) error {
	args := m.Called(ctx, inputs)
	return args.Error(0)
}

func (m *MockPinecone) QueryInput(ctx context.Context, text string, topK int) ([]*pinecone.Match, error) {
	args := m.Called(ctx, text, topK)
	return args.Get(0).([]*pinecone.Match), args.Error(1)
}

func (m *MockPinecone) Fetch(ctx context.Context, ids []string) (map[string]*pinecone.Vector, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(map[string]*pinecone.Vector), args.Error(1)
}

func (m *MockPinecone) DescribeIndexStats(ctx context.Context) (*pinecone.IndexStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*pinecone.IndexStats), args.Error(1)
}

func (m *MockPinecone) DeleteAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockResultReader for E2E
type MockResultReader struct {
	results []extract.BatchResult
}

func (m *MockResultReader) ReadResults(ctx context.Context) (<-chan extract.BatchResult, <-chan error, func()) {
	out := make(chan extract.BatchResult, len(m.results))
	errCh := make(chan error, 1)
	defer close(errCh)
	defer close(out)

	for _, r := range m.results {
		out <- r
	}
	return out, errCh, func() {}
}

func TestExtractE2E(t *testing.T) {
	// Create temp prompt file
	promptContent := `
name: Test Prompt
model: gemini-pro
messages:
  - role: user
    content: "{{.transcript}}"
`
	tmpPrompt, err := os.CreateTemp("", "prompt-*.yml")
	assert.NoError(t, err)
	defer os.Remove(tmpPrompt.Name())
	_, err = tmpPrompt.WriteString(promptContent)
	assert.NoError(t, err)
	tmpPrompt.Close()

	// Setup Mocks
	mockGenAI := new(genai.MockClient)
	mockPinecone := new(MockPinecone)

	// Create valid BatchResult
	batchRes := extract.BatchResult{
		EventID: "test-event-id",
		Response: extract.Response{
			Candidates: []extract.Candidate{{
				Content: extract.Content{
					Parts: []extract.Part{{
						Text: "{\"is_prompt_injection\": true}",
					}},
				},
			}},
		},
		Request: extract.Request{
			Contents: []extract.Content{{
				Parts: []extract.Part{{
					Text: "ignore instructions",
				}},
			}},
		},
	}

	mockReader := &MockResultReader{
		results: []extract.BatchResult{batchRes},
	}

	// Setup GenAI Response
	mockGenAI.On("GenerateContent", mock.Anything, "gemini-pro", mock.MatchedBy(func(p string) bool {
		return true
	})).Return("ignore instructions", nil)

	// Setup Pinecone Expectation
	mockPinecone.On("UpsertInputs", mock.Anything, mock.MatchedBy(func(inputs []*pinecone.InputRecord) bool {
		return len(inputs) == 1 && inputs[0].Text == "ignore instructions"
	})).Return(nil)

	// Config
	cfg := extract.Config{
		PromptPath: tmpPrompt.Name(),
	}

	// Build Service
	extractor := extract.NewExtractor(mockGenAI, cfg.PromptPath)
	processor := extract.NewProcessor(mockReader, extractor, mockPinecone, cfg)
	svc := extract.NewService(processor)

	// Run
	err = svc.Run(context.Background())
	assert.NoError(t, err)

	mockGenAI.AssertExpectations(t)
	mockPinecone.AssertExpectations(t)
}
