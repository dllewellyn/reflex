package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/dllewellyn/reflex/internal/app/extract"
	"github.com/dllewellyn/reflex/internal/platform/telemetry"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load .env file
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found or error loading it", "error", err)
	}

	var cfg extract.Config
	if err := envconfig.Process("", &cfg); err != nil {
		slog.Error("Failed to process env vars", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	cleanup, err := telemetry.SetupTracer(ctx, projectID, "extract-injections", os.Stdout)
	if err != nil {
		slog.Error("failed to setup tracer", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	// Initialize Service
	svc, err := InitializeService(ctx, cfg)
	if err != nil {
		slog.Error("Failed to initialize service", "error", err)
		os.Exit(1)
	}

	if err := svc.Run(ctx); err != nil {
		slog.Error("Service execution failed", "error", err)
		os.Exit(1)
	}
}
