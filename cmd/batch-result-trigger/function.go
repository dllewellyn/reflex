package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dllewellyn/reflex/internal/app/batch"
	"github.com/dllewellyn/reflex/internal/platform/schema"
	"github.com/google/uuid"
)

func init() {
	functions.CloudEvent("ProcessBatchResult", ProcessBatchResult)
}

// StorageObjectData contains the data from the Cloud Storage event.
type StorageObjectData struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

// GCSReader defines the interface for reading from GCS.
type GCSReader interface {
	NewReader(ctx context.Context, bucket, name string) (io.ReadCloser, error)
}

// DefaultGCSReader wraps storage.Client to satisfy GCSReader.
type DefaultGCSReader struct {
	client *storage.Client
}

func (r *DefaultGCSReader) NewReader(ctx context.Context, bucket, name string) (io.ReadCloser, error) {
	return r.client.Bucket(bucket).Object(name).NewReader(ctx)
}

// EventIngestor handles the core logic of processing files and publishing events.
type EventIngestor struct {
	gcs      GCSReader
	producer *batch.BatchEventProducer
}

var (
	ingestor *EventIngestor
	once     sync.Once
	initErr  error
)

func initializeClients(ctx context.Context) error {
	once.Do(func() {
		// Init GCS
		var err error
		gcsClient, err := storage.NewClient(ctx)
		if err != nil {
			initErr = err
			return
		}

		// Init Kafka
		bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
		topic := os.Getenv("KAFKA_TOPIC")
		apiKey := os.Getenv("KAFKA_API_KEY")
		apiSecret := os.Getenv("KAFKA_API_SECRET")

		if bootstrapServers == "" || topic == "" {
			// For testing purposes or if running in environment without Kafka, we might want to skip or mock.
			// But for production, this is an error.
			initErr = fmt.Errorf("missing kafka env vars")
			return
		}

		cfg := &kafka.ConfigMap{
			"bootstrap.servers": bootstrapServers,
			"security.protocol": "SASL_SSL",
			"sasl.mechanisms":   "PLAIN",
			"sasl.username":     apiKey,
			"sasl.password":     apiSecret,
		}

		p, err := kafka.NewProducer(cfg)
		if err != nil {
			initErr = err
			return
		}

		// Wrap in a simple adapter for BatchEventProducer
		adapter := &kafkaAdapter{p: p}
		producer := batch.NewBatchEventProducer(adapter, topic)

		ingestor = &EventIngestor{
			gcs:      &DefaultGCSReader{client: gcsClient},
			producer: producer,
		}
	})
	return initErr
}

type kafkaAdapter struct {
	p *kafka.Producer
}

func (k *kafkaAdapter) Publish(ctx context.Context, topic string, key string, msg any) error {
	val, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	deliveryChan := make(chan kafka.Event)
	err = k.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          val,
		Key:            []byte(key),
	}, deliveryChan)

	if err != nil {
		return err
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	close(deliveryChan)
	return nil
}

// ProcessBatchResult is the entry point for the Cloud Function.
func ProcessBatchResult(ctx context.Context, e event.Event) error {
	if ingestor == nil { // Allow injection for testing if already set
		if err := initializeClients(ctx); err != nil {
			slog.Error("Failed to initialize clients", "error", err)
			return err
		}
	}

	var data StorageObjectData
	if err := e.DataAs(&data); err != nil {
		slog.Error("Failed to parse event data", "error", err)
		return fmt.Errorf("event.DataAs: %v", err)
	}

	return ingestor.Process(ctx, data)
}

// Process handles the business logic of processing a GCS file.
func (i *EventIngestor) Process(ctx context.Context, data StorageObjectData) error {
	slog.Info("Processing file", "bucket", data.Bucket, "file", data.Name)

	// Open reader
	rc, err := i.gcs.NewReader(ctx, data.Bucket, data.Name)
	if err != nil {
		slog.Error("Failed to open GCS object", "error", err)
		return err
	}
	defer rc.Close()

	streamReader := batch.NewStreamReader(rc)

	count := 0
	for {
		var record schema.Record
		err := streamReader.ReadNext(&record)
		if err != nil {
			if err == io.EOF {
				break
			}
			slog.Error("Error reading stream", "error", err)
			return err
		}

		// Create Event
		evtID := uuid.New().String()
		now := time.Now()

		evt := schema.BatchResultEvent{
			EventId:   evtID,
			Timestamp: now,
			Source: schema.Source{
				Bucket: data.Bucket,
				File:   data.Name,
			},
			Record: record,
		}

		if err := i.producer.Produce(ctx, evt); err != nil {
			slog.Error("Failed to produce event", "error", err)
			return err
		}
		count++
	}

	slog.Info("Successfully processed file", "count", count)
	return nil
}
