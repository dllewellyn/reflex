package batch

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/dllewellyn/reflex/internal/platform/vertex"
)

// TriggerConfig holds configuration for the batch job trigger.
type TriggerConfig struct {
	// ProjectID is the Google Cloud Project ID.
	ProjectID string
	// Location is the GCP region (e.g., us-central1).
	Location string
	// InputURI is the GCS URI of the input JSONL file (gs://bucket/path/to/file.jsonl).
	InputURI string
	// OutputURIPrefix is the GCS URI prefix for output (gs://bucket/path/to/output/).
	OutputURIPrefix string
	// ModelID is the Vertex AI Model ID (e.g., publishers/google/models/gemini-2.5-flash).
	ModelID string
	// DisplayName is the user-friendly name for the batch job.
	DisplayName string
}

// TriggerBatchJob triggers a Vertex AI Batch Prediction job.
func TriggerBatchJob(ctx context.Context, client vertex.JobClient, cfg TriggerConfig) (*aiplatformpb.BatchPredictionJob, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", cfg.ProjectID, cfg.Location)

	req := &aiplatformpb.CreateBatchPredictionJobRequest{
		Parent: parent,
		BatchPredictionJob: &aiplatformpb.BatchPredictionJob{
			DisplayName: cfg.DisplayName,
			Model:       cfg.ModelID,
			InputConfig: &aiplatformpb.BatchPredictionJob_InputConfig{
				InstancesFormat: "jsonl",
				Source: &aiplatformpb.BatchPredictionJob_InputConfig_GcsSource{
					GcsSource: &aiplatformpb.GcsSource{
						Uris: []string{cfg.InputURI},
					},
				},
			},
			OutputConfig: &aiplatformpb.BatchPredictionJob_OutputConfig{
				PredictionsFormat: "jsonl",
				Destination: &aiplatformpb.BatchPredictionJob_OutputConfig_GcsDestination{
					GcsDestination: &aiplatformpb.GcsDestination{
						OutputUriPrefix: cfg.OutputURIPrefix,
					},
				},
			},
		},
	}

	slog.Info("Triggering Vertex Batch Job",
		"project", cfg.ProjectID,
		"location", cfg.Location,
		"model", cfg.ModelID,
		"input", cfg.InputURI,
	)

	return client.CreateBatchPredictionJob(ctx, req)
}
