package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"github.com/dllewellyn/reflex/internal/app/batch"
	"github.com/dllewellyn/reflex/internal/platform/gcs"
	"github.com/dllewellyn/reflex/internal/platform/vertex"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"

	"github.com/dllewellyn/reflex/internal/platform/telemetry"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn("Error loading .env file", "error", err)
	}

	ctx := context.Background()
	gcpProjectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	cleanup, err := telemetry.SetupTracer(ctx, gcpProjectID, "batch-job", os.Stdout)
	if err != nil {
		slog.Error("failed to setup tracer", "error", err)
		os.Exit(1)
	}
	defer cleanup()
	// Env var defaults
	// Env var defaults
	defaultProject := os.Getenv("GCP_PROJECT")
	if defaultProject == "" {
		defaultProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	defaultLocation := os.Getenv("GCP_LOCATION")
	if defaultLocation == "" {
		defaultLocation = os.Getenv("GOOGLE_CLOUD_LOCATION")
	}
	if defaultLocation == "" {
		defaultLocation = "us-central1"
	}
	defaultInput := os.Getenv("GCS_INPUT_URI")
	defaultOutput := os.Getenv("GCS_OUTPUT_PREFIX")

	// Bucket Config
	rawBucket := os.Getenv("GCS_RAW_PROMPT_BUCKET")
	stagingBucket := os.Getenv("GCS_BATCH_STAGING_BUCKET")
	processedBucket := os.Getenv("GCS_PROCESSED_PROMPT_BUCKET")
	legacyBucket := os.Getenv("GCS_BUCKET_NAME")

	// Fallbacks
	if stagingBucket == "" {
		stagingBucket = legacyBucket
	}
	if processedBucket == "" {
		processedBucket = legacyBucket
	}

	modelID := os.Getenv("MODEL_ID")

	defaultPrompt := os.Getenv("PROMPT_PATH")
	if defaultPrompt == "" {
		defaultPrompt = "prompts/security-judge.prompt.yml"
	}

	slog.Info("Configuration", "raw_bucket", rawBucket, "staging_bucket", stagingBucket, "processed_bucket", processedBucket)

	// Dynamic Input Calculation (Process Yesterday)
	if defaultInput == "" && stagingBucket != "" {
		today := time.Now()
		// Pattern: gs://<staging_bucket>/staging/<YYYY>/<MM>/<DD>/*.jsonl
		defaultInput = fmt.Sprintf("gs://%s/staging/%s/*.jsonl",
			stagingBucket,
			today.Format("2006/01/02"))
		slog.Info("No input URI provided. Auto-calculated for today using staging bucket", "uri", defaultInput)
	}

	// Dynamic Output Calculation
	if defaultOutput == "" && processedBucket != "" {
		today := time.Now()
		defaultOutput = fmt.Sprintf("gs://%s/results/%s/",
			processedBucket,
			today.Format("2006/01/02"))
	}

	projectID := flag.String("project", defaultProject, "Google Cloud Project ID")
	location := flag.String("location", defaultLocation, "Google Cloud Region")
	inputURI := flag.String("input", defaultInput, "GCS Input URI (gs://...)")
	outputURI := flag.String("output", defaultOutput, "GCS Output Prefix (gs://...)")
	promptFile := flag.String("prompt", defaultPrompt, "Path to prompt file")

	flag.Parse()

	if *projectID == "" || *inputURI == "" || *outputURI == "" {
		flag.Usage()
		slog.Error("Missing required configuration: project, input, or output")
		os.Exit(1)
	}

	// 1. Load Prompt (Validation Check)
	slog.Info("Loading prompt...", "path", *promptFile)
	prompt, err := batch.LoadPrompt(*promptFile)
	if err != nil {
		slog.Error("Failed to load prompt", "error", err)
		os.Exit(1)
	}
	slog.Info("Prompt loaded successfully", "name", prompt.Name)

	// 2. Initialize Vertex Client
	endpoint := fmt.Sprintf("%s-aiplatform.googleapis.com:443", *location)
	jobClient, err := aiplatform.NewJobClient(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		slog.Error("Failed to create Vertex Job Client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := jobClient.Close(); err != nil {
			slog.Error("Failed to close Vertex Job Client", "error", err)
		}
	}()

	client := vertex.NewClient(jobClient)

	// 3. Initialize GCS Clients
	rawGCS, err := gcs.NewClient(ctx, rawBucket)
	if err != nil {
		slog.Error("Failed to create raw GCS client", "bucket", rawBucket, "error", err)
		os.Exit(1)
	}
	defer rawGCS.Close()

	stagingGCS, err := gcs.NewClient(ctx, stagingBucket)
	if err != nil {
		slog.Error("Failed to create staging GCS client", "bucket", stagingBucket, "error", err)
		os.Exit(1)
	}
	defer stagingGCS.Close()

	// 4. Run Batch Service
	cfg := batch.Config{
		ProjectID:     *projectID,
		Location:      *location,
		StagingBucket: stagingBucket,
		OutputBucket:  processedBucket,
		ModelID:       modelID,
	}

	// Calculate target date (today by default)
	targetDate := time.Now()
	slog.Info("Running batch job", "target_date", targetDate)

	svc := batch.NewService(cfg, prompt, rawGCS, stagingGCS, client, nil) // Producer nil for now as per plan
	if err := svc.Run(ctx, targetDate); err != nil {
		slog.Error("Batch job failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Batch Job Completed Successfully")
}
