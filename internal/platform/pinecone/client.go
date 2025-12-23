package pinecone

import (
	"context"
	"fmt"

	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client is a wrapper around the Pinecone SDK.
type Client struct {
	client        *pinecone.Client
	idxConnection *pinecone.IndexConnection
}

// NewClient creates a new Pinecone client and initializes the index connection.
func NewClient(ctx context.Context, apiKey, indexHost string) (*Client, error) {
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pinecone client: %w", err)
	}

	idxConnection, err := pc.Index(pinecone.NewIndexConnParams{
		Host: indexHost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create index connection: %w", err)
	}

	return &Client{
		client:        pc,
		idxConnection: idxConnection,
	}, nil
}

// UpsertBatch upserts a batch of vectors to the configured index.
func (c *Client) UpsertBatch(ctx context.Context, vectors []*Vector) error {
	tr := otel.Tracer("pinecone-client")
	ctx, span := tr.Start(ctx, "Pinecone.UpsertBatch")
	defer span.End()

	span.SetAttributes(
		attribute.Int("pinecone.batch_size", len(vectors)),
	)

	pcVectors := make([]*pinecone.Vector, len(vectors))
	for i, v := range vectors {
		metadata, err := structpb.NewStruct(v.Metadata)
		if err != nil {
			return fmt.Errorf("failed to convert metadata to protobuf struct: %w", err)
		}
		pcVectors[i] = &pinecone.Vector{
			Id:       v.ID,
			Values:   &v.Values,
			Metadata: metadata,
		}
	}

	_, err := c.idxConnection.UpsertVectors(ctx, pcVectors)
	if err != nil {
		return fmt.Errorf("failed to upsert vectors: %w", err)
	}

	return nil
}

// UpsertInputs upserts a batch of text records to the configured index using integrated inference.
func (c *Client) UpsertInputs(ctx context.Context, inputs []*InputRecord) error {
	tr := otel.Tracer("pinecone-client")
	ctx, span := tr.Start(ctx, "Pinecone.UpsertInputs")
	defer span.End()

	span.SetAttributes(
		attribute.Int("pinecone.batch_size", len(inputs)),
	)

	// If the number of upserts > 96 it will fail, so we will throw an error if that happens
	if len(inputs) > 96 {
		return fmt.Errorf("batch size exceeds maximum allowed by Pinecone: %d", len(inputs))
	}

	// Use UpsertRecords for integrated inference
	pcRecords := make([]*pinecone.IntegratedRecord, len(inputs))
	for i, v := range inputs {
		// Ensure text is in metadata for integrated inference if needed, but for UpsertRecords
		// it expects specific fields in the map.

		record := pinecone.IntegratedRecord{
			"id":         v.ID,
			"chunk_text": "passage: " + v.Text,
		}

		// Add metadata fields to the record
		for k, val := range v.Metadata {
			record[k] = val
		}

		pcRecords[i] = &record
	}

	err := c.idxConnection.UpsertRecords(ctx, pcRecords)
	if err != nil {
		return fmt.Errorf("failed to upsert records (inference): %w", err)
	}

	return nil
}

// QueryInput queries the index using a text input for integrated inference.
func (c *Client) QueryInput(ctx context.Context, text string, topK int) ([]*Match, error) {
	tr := otel.Tracer("pinecone-client")
	ctx, span := tr.Start(ctx, "Pinecone.QueryInput")
	defer span.End()

	span.SetAttributes(
		attribute.String("pinecone.query_text", text),
		attribute.Int("pinecone.top_k", topK),
	)

	inputs := map[string]interface{}{
		"text": "query: " + text,
	}

	resp, err := c.idxConnection.SearchRecords(ctx, &pinecone.SearchRecordsRequest{
		Query: pinecone.SearchRecordsQuery{
			Inputs: &inputs,
			TopK:   int32(topK),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search records: %w", err)
	}

	matches := make([]*Match, len(resp.Result.Hits))
	for i, hit := range resp.Result.Hits {
		matches[i] = &Match{
			ID:       hit.Id,
			Score:    hit.Score,
			Metadata: hit.Fields,
		}
	}

	return matches, nil
}

// Fetch retrieves vectors by their IDs.
func (c *Client) Fetch(ctx context.Context, ids []string) (map[string]*Vector, error) {
	tr := otel.Tracer("pinecone-client")
	ctx, span := tr.Start(ctx, "Pinecone.Fetch")
	defer span.End()

	resp, err := c.idxConnection.FetchVectors(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vectors: %w", err)
	}

	vectors := make(map[string]*Vector)
	for id, vec := range resp.Vectors {
		vectors[id] = &Vector{
			ID:       vec.Id,
			Values:   *vec.Values,
			Metadata: vec.Metadata.AsMap(),
		}
	}

	return vectors, nil
}

// DescribeIndexStats retrieves statistics about the index.
func (c *Client) DescribeIndexStats(ctx context.Context) (*IndexStats, error) {
	tr := otel.Tracer("pinecone-client")
	ctx, span := tr.Start(ctx, "Pinecone.DescribeIndexStats")
	defer span.End()

	resp, err := c.idxConnection.DescribeIndexStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index stats: %w", err)
	}

	return &IndexStats{
		TotalVectorCount: resp.TotalVectorCount,
	}, nil
}

// DeleteAll deletes all vectors from the index.
func (c *Client) DeleteAll(ctx context.Context) error {
	tr := otel.Tracer("pinecone-client")
	ctx, span := tr.Start(ctx, "Pinecone.DeleteAll")
	defer span.End()

	// Use DeleteAllVectorsInNamespace with empty namespace (default)
	err := c.idxConnection.DeleteAllVectorsInNamespace(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete all vectors: %w", err)
	}

	return nil
}
