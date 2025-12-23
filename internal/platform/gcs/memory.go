package gcs

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryClient is an in-memory implementation of BlobReader and BlobWriter.
type MemoryClient struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryClient() *MemoryClient {
	return &MemoryClient{
		data: make(map[string][]byte),
	}
}

// Write writes data to the in-memory map.
func (m *MemoryClient) Write(ctx context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = data
	return nil
}

// Read reads data from the in-memory map.
func (m *MemoryClient) Read(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.data[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return data, nil
}

// ListActiveSessions returns a list of session IDs that have data for the given date.
// Assumes keys are format: raw/<session_id>/YYYY/MM/DD/HH/chunk.jsonl
func (m *MemoryClient) ListActiveSessions(ctx context.Context, date time.Time) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make(map[string]struct{})
	datePrefix := fmt.Sprintf("%d/%02d/%02d/", date.Year(), date.Month(), date.Day())

	for key := range m.data {
		if strings.HasPrefix(key, "raw/") {
			// raw/<session_id>/YYYY/MM/DD/HH/...
			parts := strings.Split(key, "/")
			if len(parts) >= 6 {
				sessionID := parts[1]
				// Check if this key matches the date
				// Key structure: raw, session, Y, M, D, H, chunk
				// Date check: parts[2]=Y, parts[3]=M, parts[4]=D
				keyDatePart := fmt.Sprintf("%s/%s/%s/", parts[2], parts[3], parts[4])
				if keyDatePart == datePrefix {
					sessions[sessionID] = struct{}{}
				}
			}
		}
	}

	var result []string
	for s := range sessions {
		result = append(result, s)
	}
	return result, nil
}

// ListSessionChunks returns all keys for a given session ID.
func (m *MemoryClient) ListSessionChunks(ctx context.Context, sessionID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var chunks []string
	prefix := fmt.Sprintf("raw/%s/", sessionID)
	for key := range m.data {
		if strings.HasPrefix(key, prefix) {
			chunks = append(chunks, key)
		}
	}
	return chunks, nil
}

// ListFiles returns a list of all object keys matching the prefix.
func (m *MemoryClient) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var files []string
	for key := range m.data {
		if strings.HasPrefix(key, prefix) {
			files = append(files, key)
		}
	}
	return files, nil
}
