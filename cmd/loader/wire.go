//go:build wireinject
// +build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	"context"
	"os"

	"github.com/dllewellyn/reflex/internal/app/loader"
	"github.com/dllewellyn/reflex/internal/platform/gcs"
	"github.com/dllewellyn/reflex/internal/platform/kafka"
	"github.com/google/wire"
)

func InitializeLoader(ctx context.Context) (*loader.Service, error) {
	wire.Build(
		loader.NewService,
		kafka.NewConsumer,
		provideGCSBucket,
		provideLoaderConfig,
		gcs.NewClient,
		wire.Bind(new(kafka.Consumer), new(*kafka.ConfluentConsumer)),
		wire.Bind(new(gcs.BlobWriter), new(*gcs.Client)),
	)
	return &loader.Service{}, nil
}

func provideGCSBucket() string {
	return os.Getenv("GCS_RAW_PROMPT_BUCKET")
}

func provideLoaderConfig() loader.Config {
	return loader.Config{
		Topic: os.Getenv("KAFKA_TOPIC"),
	}
}
