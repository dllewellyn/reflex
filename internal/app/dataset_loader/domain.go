package dataset_loader

// SourceRecord represents a raw record from the dataset source.
// It is flexible to handle varying schemas from different parquet files.
type SourceRecord map[string]interface{}

// IngestionRecord represents a record ready for ingestion into the vector store.
type IngestionRecord struct {
	ID       string                 // Unique identifier for the record
	Text     string                 // The text content to be embedded
	Label    string                 // The classification label (e.g., "injection", "safe")
	Metadata map[string]interface{} // Additional metadata to be stored
}
