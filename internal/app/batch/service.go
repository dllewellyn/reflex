package batch

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/dllewellyn/reflex/internal/platform/gcs"
	"github.com/dllewellyn/reflex/internal/platform/kafka"
	"github.com/dllewellyn/reflex/internal/platform/vertex"
)

type Config struct {
	ProjectID     string
	Location      string
	StagingBucket string
	OutputBucket  string
	ModelID       string
}

type Service struct {
	config       Config
	prompt       *Prompt
	gcsReader    gcs.BlobReader
	gcsWriter    gcs.BlobWriter
	vertexClient vertex.JobClient
	producer     kafka.Producer
}

func NewService(cfg Config, prompt *Prompt, gcsReader gcs.BlobReader, gcsWriter gcs.BlobWriter, vertexClient vertex.JobClient, producer kafka.Producer) *Service {
	return &Service{
		config:       cfg,
		prompt:       prompt,
		gcsReader:    gcsReader,
		gcsWriter:    gcsWriter,
		vertexClient: vertexClient,
		producer:     producer,
	}
}

func (s *Service) Run(ctx context.Context, targetDate time.Time) error {
	slog.Info("Starting Daily Batch Analyzer", "date", targetDate)

	// 1. List active sessions for the date
	sessions, err := s.gcsReader.ListActiveSessions(ctx, targetDate)
	if err != nil {
		return err
	}
	slog.Info("Found active sessions", "count", len(sessions))

	// 2. For each session, read full history (scaffold logic)
	processedCount := 0
	transformer := NewTransformer(s.prompt)

	for _, sessionID := range sessions {
		if err := s.processSession(ctx, sessionID, targetDate, transformer); err != nil {
			slog.Error("Failed to process session", "session_id", sessionID, "error", err)
			continue
		}
		processedCount++
	}

	if processedCount == 0 {
		slog.Info("No sessions to process")
		return nil
	}

	// 4. Trigger Batch Job
	return s.triggerBatchJob(ctx, targetDate, processedCount)
}

// processSession handles the end-to-end processing of a single session.
func (s *Service) processSession(ctx context.Context, sessionID string, date time.Time, transformer *Transformer) error {
	// 1. Reconstruct Transcript
	transcript, err := s.reconstructTranscript(ctx, sessionID)
	if err != nil {
		return err
	}
	if transcript == "" {
		return nil // Skip empty sessions
	}

	// 2. Transform to Batch Request
	data, err := transformer.CreateBatchRequest(transcript)
	if err != nil {
		return err
	}

	// 3. Upload to Staging
	return s.uploadToStaging(ctx, sessionID, date, data)
}

// reconstructTranscript reads all chunks for a session and combines them.
func (s *Service) reconstructTranscript(ctx context.Context, sessionID string) (string, error) {
	chunks, err := s.gcsReader.ListSessionChunks(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to list chunks: %w", err)
	}

	var transcriptBuilder strings.Builder
	for _, chunkKey := range chunks {
		data, err := s.gcsReader.Read(ctx, chunkKey)
		if err != nil {
			return "", fmt.Errorf("failed to read chunk %s: %w", chunkKey, err)
		}
		transcriptBuilder.Write(data)
		transcriptBuilder.WriteString("\n")
	}

	return transcriptBuilder.String(), nil
}

// uploadToStaging writes the batch input file to the staging bucket.
func (s *Service) uploadToStaging(ctx context.Context, sessionID string, date time.Time, data []byte) error {
	stagingPath := fmt.Sprintf("staging/%s/%s.jsonl", date.Format("2006/01/02"), sessionID)
	if err := s.gcsWriter.Write(ctx, stagingPath, data); err != nil {
		return fmt.Errorf("failed to upload to %s: %w", stagingPath, err)
	}
	return nil
}

// triggerBatchJob submits the job to Vertex AI.
func (s *Service) triggerBatchJob(ctx context.Context, date time.Time, sessionCount int) error {
	// Use wildcard to include all session files for this date
	inputURI := fmt.Sprintf("gs://%s/staging/%s/*.jsonl", s.config.StagingBucket, date.Format("2006/01/02"))
	outputURI := fmt.Sprintf("gs://%s/results/%s/", s.config.OutputBucket, date.Format("2006/01/02"))
	jobName := fmt.Sprintf("security-judge-%s", date.Format("2006-01-02"))

	slog.Info("Triggering batch job", "input_pattern", inputURI, "session_count", sessionCount)

	resp, err := s.vertexClient.CreateBatchPredictionJob(ctx, &aiplatformpb.CreateBatchPredictionJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", s.config.ProjectID, s.config.Location),
		BatchPredictionJob: &aiplatformpb.BatchPredictionJob{
			DisplayName: jobName,
			Model:       s.config.ModelID,
			InputConfig: &aiplatformpb.BatchPredictionJob_InputConfig{
				InstancesFormat: "jsonl",
				Source: &aiplatformpb.BatchPredictionJob_InputConfig_GcsSource{
					GcsSource: &aiplatformpb.GcsSource{
						Uris: []string{inputURI},
					},
				},
			},
			OutputConfig: &aiplatformpb.BatchPredictionJob_OutputConfig{
				PredictionsFormat: "jsonl",
				Destination: &aiplatformpb.BatchPredictionJob_OutputConfig_GcsDestination{
					GcsDestination: &aiplatformpb.GcsDestination{
						OutputUriPrefix: outputURI,
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create batch job: %w", err)
	}

	slog.Info("Batch Job Triggered Successfully", "job_name", resp.Name, "state", resp.State)
	return nil
}
