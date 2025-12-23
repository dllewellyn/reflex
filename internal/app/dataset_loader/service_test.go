package dataset_loader_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/dllewellyn/reflex/internal/app/dataset_loader"
	"github.com/dllewellyn/reflex/internal/platform/huggingface"
	"github.com/dllewellyn/reflex/internal/platform/pinecone"
)

// MockDownloader
type MockDownloader struct {
	reader huggingface.DatasetReader
	err    error
}

func (m *MockDownloader) DownloadAndRead(ctx context.Context, datasetID, split string) (huggingface.DatasetReader, error) {
	return m.reader, m.err
}

// MockReader
type MockReader struct {
	rowsToReturn []map[string]interface{}
	errToReturn  error
	offset       int
}

func (m *MockReader) Read(rows []map[string]interface{}) (int, error) {
	if m.offset >= len(m.rowsToReturn) {
		if m.errToReturn != nil {
			return 0, m.errToReturn
		}
		return 0, io.EOF
	}

	n := 0
	for i := 0; i < len(rows) && m.offset < len(m.rowsToReturn); i++ {
		rows[i] = m.rowsToReturn[m.offset]
		m.offset++
		n++
	}

	return n, nil
}

func (m *MockReader) Close() error {
	return nil
}

// MockVectorStore
type MockVectorStore struct {
	UpsertInputsFunc func(ctx context.Context, inputs []*pinecone.InputRecord) error
	FetchFunc        func(ctx context.Context, ids []string) (map[string]*pinecone.Vector, error)
	StatsFunc        func(ctx context.Context) (*pinecone.IndexStats, error)
	UpsertCount      int
}

func (m *MockVectorStore) UpsertBatch(ctx context.Context, vectors []*pinecone.Vector) error {
	return nil
}
func (m *MockVectorStore) UpsertInputs(ctx context.Context, inputs []*pinecone.InputRecord) error {
	if m.UpsertInputsFunc != nil {
		return m.UpsertInputsFunc(ctx, inputs)
	}
	return nil
}
func (m *MockVectorStore) QueryInput(ctx context.Context, text string, topK int) ([]*pinecone.Match, error) {
	return nil, nil
}
func (m *MockVectorStore) Fetch(ctx context.Context, ids []string) (map[string]*pinecone.Vector, error) {
	if m.FetchFunc != nil {
		return m.FetchFunc(ctx, ids)
	}
	return map[string]*pinecone.Vector{}, nil
}
func (m *MockVectorStore) DescribeIndexStats(ctx context.Context) (*pinecone.IndexStats, error) {
	if m.StatsFunc != nil {
		return m.StatsFunc(ctx)
	}
	return &pinecone.IndexStats{TotalVectorCount: 0}, nil
}
func (m *MockVectorStore) DeleteAll(ctx context.Context) error {
	return nil
}

func TestService_Run_ReportErrorOnReadFailure(t *testing.T) {
	// Scenario 1: Read fails immediately with non-EOF error.
	mockReader := &MockReader{
		rowsToReturn: []map[string]interface{}{},
		errToReturn:  errors.New("corruption error"),
	}
	mockDownloader := &MockDownloader{reader: mockReader}
	svc := dataset_loader.NewService(dataset_loader.Config{BatchSize: 10}, mockDownloader, &MockVectorStore{})

	err := svc.Run(context.Background())
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "error reading dataset: corruption error" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestService_Run_Success(t *testing.T) {
	// Scenario 2: Read succeeds.
	mockReader := &MockReader{
		rowsToReturn: []map[string]interface{}{
			{"text": "hello", "label": 0},
		},
		errToReturn: nil,
	}
	mockDownloader := &MockDownloader{reader: mockReader}
	svc := dataset_loader.NewService(dataset_loader.Config{BatchSize: 10, HFTextCol: "text"}, mockDownloader, &MockVectorStore{})

	err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestService_Run_DeduplicationAndVerification(t *testing.T) {
	// Scenario 3: Deduplication and Verification
	// 2 rows: one exists, one is new.
	mockReader := &MockReader{
		rowsToReturn: []map[string]interface{}{
			{"text": "existing", "label": 0},
			{"text": "new", "label": 1},
		},
		errToReturn: nil,
	}
	mockDownloader := &MockDownloader{reader: mockReader}

	// We can pre-calculate the ID for "existing" if we wanted to be precise,
	// but for the mock FetchFunc we can just return one generic ID to simulate existence.
	// However, the service calculates SHA256 IDs.
	// Let's rely on the service to call Fetch and we return one match.

	Store := &MockVectorStore{
		StatsFunc: func(ctx context.Context) (*pinecone.IndexStats, error) {
			return &pinecone.IndexStats{TotalVectorCount: 10}, nil
		},
		FetchFunc: func(ctx context.Context, ids []string) (map[string]*pinecone.Vector, error) {
			// Simulate that the first ID exists
			if len(ids) > 0 {
				return map[string]*pinecone.Vector{
					ids[0]: {}, // First ID exists
				}, nil
			}
			return nil, nil
		},
		UpsertInputsFunc: func(ctx context.Context, inputs []*pinecone.InputRecord) error {
			if len(inputs) != 1 {
				t.Errorf("Expected 1 input to be upserted (the new one), got %d", len(inputs))
			}
			return nil
		},
	}

	svc := dataset_loader.NewService(dataset_loader.Config{BatchSize: 10, HFTextCol: "text"}, mockDownloader, Store)

	err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}
