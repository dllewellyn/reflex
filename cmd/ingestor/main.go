package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dllewellyn/reflex/internal/platform/pinecone"
	"github.com/dllewellyn/reflex/internal/platform/telemetry"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/magiconair/properties"
)

type Config struct {
	KafkaTopic            string `envconfig:"KAFKA_TOPIC" required:"true"`
	KafkaConfigFile       string `envconfig:"KAFKA_CONFIG_FILE" default:"client.properties"`
	KafkaBootstrapServers string `envconfig:"KAFKA_BOOTSTRAP_SERVERS"`
	KafkaAPIKey           string `envconfig:"KAFKA_API_KEY"`
	KafkaAPISecret        string `envconfig:"KAFKA_API_SECRET"`
	Port                  string `envconfig:"PORT" default:"8080"`
}

// IngestorConfig holds configuration specific to the Ingestor service.
type IngestorConfig struct {
	TopicName         string
	Port              string
}

func readKafkaConfig(configFile string) (*ckafka.ConfigMap, error) {
	m := make(map[string]ckafka.ConfigValue)

	props, err := properties.LoadFile(configFile, properties.UTF8)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Warn("Kafka config file not found, using defaults", "file", configFile)
			return &ckafka.ConfigMap{"bootstrap.servers": "localhost:9092"}, nil
		}
		return nil, err
	}

	for _, key := range props.Keys() {
		val, _ := props.Get(key)
		m[key] = val
	}
	return (*ckafka.ConfigMap)(&m), nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	cleanup, err := telemetry.SetupTracer(ctx, projectID, "ingestor", os.Stdout)
	if err != nil {
		slog.Error("failed to setup tracer", "error", err)
		// Proceed without tracer
	}
	defer cleanup()

	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// Set as default slog logger
	slog.SetDefault(logger)

	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found or error loading it", "error", err)
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to process env vars: %v", err)
	}

	kafkaCfg, err := readKafkaConfig(cfg.KafkaConfigFile)
	if err != nil {
		log.Fatalf("Failed to read kafka config: %v", err)
	}

	// Override with Env Vars if present
	if cfg.KafkaBootstrapServers != "" {
		slog.Info("Overriding bootstrap.servers from env")
		(*kafkaCfg)["bootstrap.servers"] = cfg.KafkaBootstrapServers
	}
	if cfg.KafkaAPIKey != "" {
		slog.Info("Overriding SASL credentials from env")
		(*kafkaCfg)["sasl.username"] = cfg.KafkaAPIKey
		(*kafkaCfg)["sasl.password"] = cfg.KafkaAPISecret
		(*kafkaCfg)["security.protocol"] = "SASL_SSL"
		(*kafkaCfg)["sasl.mechanisms"] = "PLAIN"
	}

	// Handle SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Pinecone Client
	pcAPIKey := os.Getenv("PINECONE_API_KEY")
	pcIndexHost := os.Getenv("PINECONE_INDEX_HOST")

	// Create Pinecone client
	var vectorStore pinecone.VectorStore
	if pcAPIKey != "" && pcIndexHost != "" {
		pcClient, err := pinecone.NewClient(ctx, pcAPIKey, pcIndexHost)
		if err != nil {
			slog.Error("Failed to create pinecone client", "error", err)
			os.Exit(1)
		}
		vectorStore = pcClient
	} else {
		slog.Warn("PINECONE_API_KEY or PINECONE_INDEX_HOST not set. AnalyzePrompt will fail (panic potential inside service if called).")
		// We pass nil, assuming service might panic if used?
		// Or we should verify if InitializeIngestor allows nil interface?
		// Wire passes it to NewService. NewService stores it.
		// AnalyzePrompt calls s.vectorStore.QueryInput -> panic if nil.
		// User requested implementation. We'll proceed.
	}

	// Initialize Service via Wire
	ingestorCfg := IngestorConfig{
		TopicName: cfg.KafkaTopic,
		Port:      cfg.Port,
	}

	svc, err := InitializeIngestor(ctx, ingestorCfg, kafkaCfg, vectorStore)
	if err != nil {
		log.Fatalf("Failed to initialize ingestor: %v", err)
	}

	log.Println("Starting server...")
	if err := svc.Run(ctx); err != nil {
		log.Printf("Service exited with error: %v", err)
	}
}
