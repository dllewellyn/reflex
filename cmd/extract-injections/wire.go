//go:build wireinject
// +build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	"context"

	"github.com/dllewellyn/reflex/internal/app/extract"
	"github.com/dllewellyn/reflex/internal/platform/genai"
	"github.com/dllewellyn/reflex/internal/platform/pinecone"
	"github.com/google/wire"
)

func InitializeService(ctx context.Context, cfg extract.Config) (*extract.Service, error) {
	wire.Build(
		extract.NewService,
		extract.NewProcessor,
		extract.NewExtractor,

		provideResultReader, // New provider
		wire.Bind(new(genai.ClientInterface), new(*genai.Client)),
		wire.Bind(new(pinecone.VectorStore), new(*pinecone.Client)),

		provideGenAIClient,
		providePineconeClient,
		provideExtractorPromptFile,
	)
	return nil, nil
}

func provideResultReader(cfg extract.Config) extract.ResultReader {
	return extract.NewKafkaResultReader(cfg)
}

func provideGenAIClient(ctx context.Context, cfg extract.Config) (*genai.Client, error) {
	return genai.NewClient(ctx, cfg.GCPProjectID, cfg.GCPLocation)
}

func providePineconeClient(ctx context.Context, cfg extract.Config) (*pinecone.Client, error) {
	return pinecone.NewClient(ctx, cfg.PineconeAPIKey, cfg.PineconeIndexHost)
}

func provideExtractorPromptFile(cfg extract.Config) string {
	return cfg.PromptPath
}
