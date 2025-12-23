package gcs

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/api/iterator"
)

type Client struct {
	client     *storage.Client
	bucketName string
}

func NewClient(ctx context.Context, bucketName string) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	return &Client{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// Write writes data to a GCS object at the specified key.
func (w *Client) Write(ctx context.Context, key string, data []byte) error {
	tr := otel.Tracer("gcs-writer")
	ctx, span := tr.Start(ctx, "GCS.Write")
	defer span.End()

	span.SetAttributes(
		attribute.String("gcs.bucket", w.bucketName),
		attribute.String("gcs.key", key),
	)

	wc := w.client.Bucket(w.bucketName).Object(key).NewWriter(ctx)

	if _, err := wc.Write(data); err != nil {
		_ = wc.Close()
		return fmt.Errorf("failed to write data to GCS object %s: %w", key, err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close/commit GCS object %s: %w", key, err)
	}

	return nil
}

// Read reads the content of a GCS object at the specified key.
func (w *Client) Read(ctx context.Context, key string) ([]byte, error) {
	tr := otel.Tracer("gcs-writer")
	ctx, span := tr.Start(ctx, "GCS.Read")
	defer span.End()

	span.SetAttributes(
		attribute.String("gcs.bucket", w.bucketName),
		attribute.String("gcs.key", key),
	)

	rc, err := w.client.Bucket(w.bucketName).Object(key).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader for %s: %w", key, err)
	}
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from %s: %w", key, err)
	}
	return data, nil
}

// ListActiveSessions returns a list of session IDs that had activity on the specified date.
func (w *Client) ListActiveSessions(ctx context.Context, date time.Time) ([]string, error) {
	// Strategy: List all objects in the bucket, but that's too many.
	// Structure: raw/<session_id>/YYYY/MM/DD/HH/chunk.jsonl
	// We want sessions where raw/<session_id>/<target_date> exists.
	// This requires scanning.

	// For now, I will implement a basic scan of top-level prefixes under 'raw/'.
	// And then for each, check if the date-specific path exists.
	// This is O(Sessions).

	var sessions []string
	it := w.client.Bucket(w.bucketName).Objects(ctx, &storage.Query{
		Prefix:    "raw/",
		Delimiter: "/",
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		// attrs.Prefix will be "raw/<session_id>/"
		if attrs.Prefix != "" {
			// Check if this session has data for the date
			// Date path: raw/<session_id>/YYYY/MM/DD
			sessionPath := attrs.Prefix
			datePath := fmt.Sprintf("%s%d/%02d/%02d/", sessionPath, date.Year(), date.Month(), date.Day())

			// Check existence by listing with this prefix and limiting to 1 result
			dateIt := w.client.Bucket(w.bucketName).Objects(ctx, &storage.Query{
				Prefix: datePath,
				// We don't need delimiter, just want to know if anything exists
			})
			_, err := dateIt.Next()
			if err == nil {
				// Found at least one object
				// Extract session ID from "raw/<session_id>/"
				// "raw/" is 4 chars.
				// sessionID is attrs.Prefix[4 : len(attrs.Prefix)-1]
				sessionID := sessionPath[4 : len(sessionPath)-1]
				sessions = append(sessions, sessionID)
			} else if err != iterator.Done {
				// Real error
				// log it but maybe continue?
				fmt.Printf("Error checking date path %s: %v\n", datePath, err)
			}
		}
	}
	return sessions, nil
}

func (w *Client) ListSessionChunks(ctx context.Context, sessionID string) ([]string, error) {
	prefix := fmt.Sprintf("raw/%s/", sessionID)
	it := w.client.Bucket(w.bucketName).Objects(ctx, &storage.Query{
		Prefix: prefix,
	})

	var chunks []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, attrs.Name)
	}
	return chunks, nil
}

// ListFiles returns a list of all object keys matching the prefix.
func (w *Client) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	it := w.client.Bucket(w.bucketName).Objects(ctx, &storage.Query{
		Prefix: prefix,
	})

	var files []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, attrs.Name)
	}
	return files, nil
}

func (w *Client) Close() error {
	return w.client.Close()
}
