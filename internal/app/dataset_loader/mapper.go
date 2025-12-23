package dataset_loader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// MapRowToIngestionRecord maps a source record (row) to an IngestionRecord.
func MapRowToIngestionRecord(row SourceRecord, config Config) (*IngestionRecord, error) {
	textVal, ok := row[config.HFTextCol]
	if !ok {
		return nil, fmt.Errorf("missing text column: %s", config.HFTextCol)
	}
	text, ok := textVal.(string)
	if !ok {
		// Attempt to convert to string if it's not
		if strVal, ok := textVal.(fmt.Stringer); ok {
			text = strVal.String()
		} else {
			text = fmt.Sprintf("%v", textVal)
		}
	}

	labelVal, ok := row[config.HFLabelCol]
	label := ""
	if ok {
		if l, ok := labelVal.(string); ok {
			label = l
		} else {
			label = fmt.Sprintf("%v", labelVal)
		}
	}

	// Generate a deterministic ID based on the content (Phase 4 / T013 requirement brought forward for robustness)
	// Using SHA256 of the text content to ensure idempotency.
	hash := sha256.Sum256([]byte(text))
	id := hex.EncodeToString(hash[:])

	// Filter metadata: include everything except text and label to avoid duplication,
	// or just include specific fields. For now, we'll keep everything else as metadata.
	metadata := make(map[string]interface{})
	for k, v := range row {
		if k != config.HFTextCol {
			metadata[k] = v
		}
	}
	// Explicitly add label to metadata if needed, but it's already in the struct.
	// The Pinecone client expects metadata in the Vector struct.
	// We will ensure the Label is part of the metadata passed to Pinecone.
	metadata["label"] = label
	metadata["text"] = text // Often useful to have the text in metadata for retrieval

	return &IngestionRecord{
		ID:       id,
		Text:     text,
		Label:    label,
		Metadata: metadata,
	}, nil
}
