package gcs

import (
	"context"
	"time"
)

// BlobWriter defines the interface for writing blobs to GCS.
type BlobWriter interface {
	// Write writes data to a GCS object at the specified key.
	Write(ctx context.Context, key string, data []byte) error
}

// BlobReader defines the interface for reading blobs from GCS.
type BlobReader interface {
	// ListActiveSessions returns a list of session IDs that had activity on the specified date.
	ListActiveSessions(ctx context.Context, date time.Time) ([]string, error)

	// ListSessionChunks returns a list of all object keys (chunks) for a given session ID,
	// spanning all dates.
	ListSessionChunks(ctx context.Context, sessionID string) ([]string, error)

	// ListFiles returns a list of all object keys matching the prefix.
	ListFiles(ctx context.Context, prefix string) ([]string, error)

	// Read reads the content of a GCS object at the specified key.
	Read(ctx context.Context, key string) ([]byte, error)
}
