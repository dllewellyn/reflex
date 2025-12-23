package vertex

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/dllewellyn/reflex/internal/platform/schema"
	"github.com/google/uuid"
	"github.com/googleapis/gax-go/v2"
)

// MemoryClient is an in-memory implementation of JobClient.
type MemoryClient struct {
	jobs            map[string]*schema.VertexBatchOutput
	submittedInputs []schema.VertexBatchInput
	createdJobs     []*aiplatformpb.CreateBatchPredictionJobRequest
}

func NewMemoryClient() *MemoryClient {
	return &MemoryClient{
		jobs:            make(map[string]*schema.VertexBatchOutput),
		submittedInputs: make([]schema.VertexBatchInput, 0),
		createdJobs:     make([]*aiplatformpb.CreateBatchPredictionJobRequest, 0),
	}
}

// SetJobResult configures a result to be returned for a specific job ID (for testing).
func (m *MemoryClient) SetJobResult(jobID string, result *schema.VertexBatchOutput) {
	m.jobs[jobID] = result
}

// GetCreatedJobs returns the list of job creation requests (for testing).
func (m *MemoryClient) GetCreatedJobs() []*aiplatformpb.CreateBatchPredictionJobRequest {
	return m.createdJobs
}

func (m *MemoryClient) CreateBatchPredictionJob(ctx context.Context, req *aiplatformpb.CreateBatchPredictionJobRequest, opts ...gax.CallOption) (*aiplatformpb.BatchPredictionJob, error) {
	jobID := uuid.New().String()
	m.createdJobs = append(m.createdJobs, req)

	slog.Info("MemoryClient: Created Batch Prediction Job", "jobID", jobID, "model", req.Parent)

	return &aiplatformpb.BatchPredictionJob{
		Name:  fmt.Sprintf("%s/batchPredictionJobs/%s", req.Parent, jobID),
		State: aiplatformpb.JobState_JOB_STATE_PENDING,
	}, nil
}

// Deprecated methods to keep potential compatibility or for gradual migration if needed by other tests,
// though the interface check failed on CreateBatchPredictionJob.
// I will keep SubmitJob and GetJobResult for now if they are used internally by other parts not checked by the compiler for interface satisfaction yet.
// But Service expects JobClient which ONLY has CreateBatchPredictionJob.
// So these methods are not needed for JobClient interface satisfaction, but maybe for logic I haven't seen.
// Given the build error was "missing method CreateBatchPredictionJob", adding it fixes that.
