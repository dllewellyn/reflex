package dataset_loader

// Config holds the configuration for the dataset loader service.
type Config struct {
	// HuggingFace Configuration
	HFDatasetID string `envconfig:"HF_DATASET_ID" default:"deepset/prompt-injections"`
	HFSplit     string `envconfig:"HF_SPLIT" default:"train"`
	HFTextCol   string `envconfig:"HF_TEXT_COL" default:"text"`
	HFLabelCol  string `envconfig:"HF_LABEL_COL" default:"label"`
	HFFilterCol string `envconfig:"HF_FILTER_COL"`
	HFFilterVal string `envconfig:"HF_FILTER_VAL"`

	// Pinecone Configuration
	PineconeAPIKey    string `envconfig:"PINECONE_API_KEY" required:"true"`
	PineconeIndexHost string `envconfig:"PINECONE_INDEX_HOST" required:"true"`

	// Processing Configuration
	BatchSize       int `envconfig:"BATCH_SIZE" default:"96"`
	VectorDimension int `envconfig:"VECTOR_DIMENSION" default:"1024"`
}
