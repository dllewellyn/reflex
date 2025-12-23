package pinecone

import "context"

// Vector represents a vector to be stored.
type Vector struct {
	ID       string
	Values   []float32
	Metadata map[string]interface{}
}

// InputRecord represents a record to be embedded and stored by Pinecone.
type InputRecord struct {
	ID       string
	Text     string
	Metadata map[string]interface{}
}

// Match represents a single query result.
type Match struct {
	ID       string
	Score    float32
	Metadata map[string]interface{}
}

// IndexStats represents statistics about the index.
type IndexStats struct {
	TotalVectorCount uint32
}

// VectorStore defines the interface for vector storage operations.
type VectorStore interface {
	UpsertBatch(ctx context.Context, vectors []*Vector) error
	UpsertInputs(ctx context.Context, inputs []*InputRecord) error
	QueryInput(ctx context.Context, text string, topK int) ([]*Match, error)
	Fetch(ctx context.Context, ids []string) (map[string]*Vector, error)
	DeleteAll(ctx context.Context) error
	DescribeIndexStats(ctx context.Context) (*IndexStats, error)
}
