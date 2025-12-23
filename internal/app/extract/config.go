package extract

type Config struct {
	GCPProjectID string `envconfig:"GCP_PROJECT_ID" required:"true"`
	GCPLocation  string `envconfig:"GCP_LOCATION" default:"us-central1"`

	PromptPath            string `envconfig:"PROMPT_PATH" default:"prompts/extract-injection.prompt.yml"`
	PineconeAPIKey        string `envconfig:"PINECONE_API_KEY" required:"true"`
	PineconeIndexHost     string `envconfig:"PINECONE_INDEX_HOST" required:"true"`
	KafkaBootstrapServers string `envconfig:"KAFKA_BOOTSTRAP_SERVERS"`
	KafkaTopic            string `envconfig:"KAFKA_TOPIC_BATCH_RESULTS"`
	KafkaGroupID          string `envconfig:"KAFKA_CONSUMER_GROUP_ID" default:"extract-injections-group"`
	KafkaAPIKey           string `envconfig:"KAFKA_API_KEY"`
	KafkaAPISecret        string `envconfig:"KAFKA_API_SECRET"`
	IdleTimeoutSeconds    int    `envconfig:"IDLE_TIMEOUT_SECONDS" default:"30"`
	DryRun                bool   `envconfig:"DRY_RUN" default:"false"`
}
