//go:build wireinject
// +build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dllewellyn/reflex/internal/app/ingestor"
	kafkaPlatform "github.com/dllewellyn/reflex/internal/platform/kafka"
	"github.com/dllewellyn/reflex/internal/platform/pinecone"
	"github.com/google/wire"
)

func InitializeIngestor(ctx context.Context, cfg IngestorConfig, kafkaCfg *kafka.ConfigMap, vectorStore pinecone.VectorStore) (*ingestor.Service, error) {
	wire.Build(
		kafkaPlatform.NewProducer,
		provideServiceConfig,
		// Bind implementations to interfaces
		wire.Bind(new(kafkaPlatform.Producer), new(*kafkaPlatform.ConfluentProducer)),
		ingestor.NewService,
	)
	return &ingestor.Service{}, nil
}

func provideServiceConfig(cfg IngestorConfig) ingestor.Config {
	return ingestor.Config{
		TopicName: cfg.TopicName,
		Port:      cfg.Port,
	}
}
