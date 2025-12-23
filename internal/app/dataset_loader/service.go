package dataset_loader

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/dllewellyn/reflex/internal/platform/huggingface"
	"github.com/dllewellyn/reflex/internal/platform/pinecone"
)

// Downloader defines the interface for downloading and reading datasets.
type Downloader interface {
	DownloadAndRead(ctx context.Context, datasetID, split string) (huggingface.DatasetReader, error)
}

// Service implements the dataset loader logic.
type Service struct {
	config      Config
	hfClient    Downloader
	vectorStore pinecone.VectorStore
}

// NewService creates a new instance of the dataset loader service.
func NewService(cfg Config, hfClient Downloader, vectorStore pinecone.VectorStore) *Service {
	return &Service{
		config:      cfg,
		hfClient:    hfClient,
		vectorStore: vectorStore,
	}
}

// Run executes the dataset loading process.
func (s *Service) Run(ctx context.Context) error {
	log.Printf("Starting dataset loader service with config: %+v\n", s.config)

	// Get initial index stats
	initialStats, err := s.vectorStore.DescribeIndexStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get initial index stats: %w", err)
	}
	log.Printf("Initial index record count: %d", initialStats.TotalVectorCount)

	// 1. Download and open the dataset file (Parquet, JSONL, or JSON)
	log.Printf("Downloading dataset %s (split: %s)...", s.config.HFDatasetID, s.config.HFSplit)
	reader, err := s.hfClient.DownloadAndRead(ctx, s.config.HFDatasetID, s.config.HFSplit)
	if err != nil {
		return fmt.Errorf("failed to download dataset: %w", err)
	}
	defer reader.Close()

	// 2. Process rows in batches
	batchSize := s.config.BatchSize
	rows := make([]map[string]interface{}, batchSize)

	totalProcessed := 0
	totalUpserted := 0
	totalSkipped := 0

	for {
		n, err := reader.Read(rows)
		if n > 0 {
			// Process the batch
			inputs := make([]*pinecone.InputRecord, 0, n)
			inputIDs := make([]string, 0, n)
			for i := 0; i < n; i++ {
				row := rows[i]

				// Apply filter if configured
				if s.config.HFFilterCol != "" {
					val, ok := row[s.config.HFFilterCol]
					if !ok {
						// Filter column missing, treat as mismatch/skip? Or error?
						// Let's safe skip and log widely if needed, but for now just skip.
						totalSkipped++
						continue
					}

					// Convert to string for comparison
					strVal := fmt.Sprintf("%v", val)
					if strVal != s.config.HFFilterVal {
						totalSkipped++
						continue
					}
				}

				record, err := MapRowToIngestionRecord(SourceRecord(row), s.config)
				if err != nil {
					log.Printf("Warning: skipping row due to error: %v", err)
					continue
				}

				inputs = append(inputs, &pinecone.InputRecord{
					ID:       record.ID,
					Text:     record.Text,
					Metadata: record.Metadata,
				})
				inputIDs = append(inputIDs, record.ID)
			}

			// Deduplication: Check for existing records
			existingVectors, err := s.vectorStore.Fetch(ctx, inputIDs)
			if err != nil {
				return fmt.Errorf("failed to check for existing records: %w", err)
			}

			newInputs := make([]*pinecone.InputRecord, 0, len(inputs))
			for _, input := range inputs {
				if _, exists := existingVectors[input.ID]; !exists {
					newInputs = append(newInputs, input)
				} else {
					totalSkipped++
				}
			}

			// Upsert batch
			if len(newInputs) > 0 {
				log.Printf("Upserting batch of %d records (skipped %d duplicates)...", len(newInputs), len(inputs)-len(newInputs))
				if err := s.vectorStore.UpsertInputs(ctx, newInputs); err != nil {
					return fmt.Errorf("failed to upsert batch: %w", err)
				}
				totalUpserted += len(newInputs)

				// Immediate verification check for the first batch or periodically
				if totalUpserted == len(newInputs) {
					// Check a few IDs to see if they exist
					checkIDs := []string{newInputs[0].ID}
					found, err := s.vectorStore.Fetch(ctx, checkIDs)
					if err != nil {
						log.Printf("Warning: failed to verify upsert: %v", err)
					} else if len(found) == 0 {
						log.Printf("Warning: Immediate fetch failed to find record %s. Index might be eventually consistent.", checkIDs[0])
					} else {
						log.Printf("Verified: Record %s exists.", checkIDs[0])
					}
				}

			} else {
				log.Printf("Skipped all %d records in batch as duplicates.", len(inputs))
			}
			totalProcessed += len(inputs)
			log.Printf("Processed %d records so far (upserted: %d, skipped: %d)", totalProcessed, totalUpserted, totalSkipped)
		}
		if err != nil {
			if err == context.Canceled {
				return ctx.Err()
			}
			if err == io.EOF {
				break
			}
			if err == context.DeadlineExceeded {
				return err
			}

			// Return any other error
			return fmt.Errorf("error reading dataset: %w", err)
		}
	}

	log.Printf("Dataset ingestion complete. Total processed: %d, Upserted: %d, Skipped: %d", totalProcessed, totalUpserted, totalSkipped)

	// Verification with delay
	log.Println("Waiting 5 seconds for index consistency before final verification...")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
	}

	finalStats, err := s.vectorStore.DescribeIndexStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get final index stats: %w", err)
	}
	log.Printf("Final index record count: %d", finalStats.TotalVectorCount)

	expectedCount := int(initialStats.TotalVectorCount) + totalUpserted
	if int(finalStats.TotalVectorCount) != expectedCount {
		log.Printf("Warning: Verification mismatch. Expected %d records, found %d. (Note: Index updates might be eventually consistent)", expectedCount, finalStats.TotalVectorCount)
	} else {
		log.Printf("Verification successful: Index count matches expected count.")
	}

	return nil
}

// DeleteAll deletes all items from the vector storage.
func (s *Service) DeleteAll(ctx context.Context) error {
	log.Printf("Deleting ALL items from vector storage...")
	if err := s.vectorStore.DeleteAll(ctx); err != nil {
		return fmt.Errorf("dataset loader service failed to delete all items: %w", err)
	}
	log.Printf("Successfully requested deletion of all items.")
	return nil
}
