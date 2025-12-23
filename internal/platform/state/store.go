package state

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	collectionName = "ingestor_state"
	docName        = "checkpoint"
	fieldLastRun   = "last_run"
)

// Store defines the interface for state management.
type Store interface {
	GetLastRun(ctx context.Context) (time.Time, error)
	SetLastRun(ctx context.Context, t time.Time) error
	Close() error
}

type firestoreStore struct {
	client    *firestore.Client
	projectID string
}

// NewFirestoreStore creates a new state store backed by Firestore.
// It assumes ADC (Application Default Credentials) are set up.
func NewFirestoreStore(ctx context.Context, projectID string) (Store, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}
	return &firestoreStore{client: client, projectID: projectID}, nil
}

func (s *firestoreStore) GetLastRun(ctx context.Context) (time.Time, error) {
	doc, err := s.client.Collection(collectionName).Doc(docName).Get(ctx)
	if err != nil {
		// Strictly check for NotFound. Any other error is fatal.
		if status.Code(err) == codes.NotFound {
			// If not found, it means this is the first run. Return zero time.
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to get checkpoint: %w", err)
	}

	if t, ok := doc.Data()[fieldLastRun].(time.Time); ok {
		return t, nil
	}

	// Field missing or invalid type?
	// This implies data corruption or schema change.
	// Returning zero time risks re-ingestion, but failing risks getting stuck.
	// For safety against data loss/duplication, we should probably fail or log loudly.
	// For this PoC, we'll treat it as a fresh start (zero time) but ideally we'd alert.
	return time.Time{}, nil
}

func (s *firestoreStore) SetLastRun(ctx context.Context, t time.Time) error {
	_, err := s.client.Collection(collectionName).Doc(docName).Set(ctx, map[string]interface{}{
		fieldLastRun: t,
	})
	if err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}
	return nil
}

func (s *firestoreStore) Close() error {
	return s.client.Close()
}
