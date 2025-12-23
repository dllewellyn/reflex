package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/dllewellyn/reflex/internal/platform/telemetry"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if os.Getenv("KAFKA_CONSUMER_GROUP_ID") == "" {
		os.Setenv("KAFKA_CONSUMER_GROUP_ID", "loader-consumer")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	cleanup, err := telemetry.SetupTracer(ctx, projectID, "loader", os.Stdout)
	if err != nil {
		slog.Error("failed to setup tracer", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	slog.Info("Starting Hourly Loader Job...")

	svc, err := InitializeLoader(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize loader: %v", err)
	}

	// Run the batch load process
	if err := svc.RunOnce(ctx); err != nil {
		log.Fatalf("Loader job failed: %v", err)
	}

	log.Println("Loader Job completed successfully.")
}
