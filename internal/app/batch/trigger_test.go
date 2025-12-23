package batch

import (
	"context"
	"testing"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/dllewellyn/reflex/internal/platform/vertex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTriggerBatchJob(t *testing.T) {
	mockClient := &vertex.MockJobClient{}

	// Expectations
	expectedJobName := "projects/my-project/locations/us-central1/batchPredictionJobs/job-123"
	mockClient.On("CreateBatchPredictionJob",
		mock.Anything,
		mock.AnythingOfType("*aiplatformpb.CreateBatchPredictionJobRequest"),
		mock.Anything,
	).Return(&aiplatformpb.BatchPredictionJob{
		Name:  expectedJobName,
		State: aiplatformpb.JobState_JOB_STATE_PENDING,
	}, nil)

	cfg := TriggerConfig{
		ProjectID:       "my-project",
		Location:        "us-central1",
		InputURI:        "gs://bucket/input.jsonl",
		OutputURIPrefix: "gs://bucket/output/",
		ModelID:         "gemini-2.5-flash",
		DisplayName:     "test-job",
	}

	job, err := TriggerBatchJob(context.Background(), mockClient, cfg)
	require.NoError(t, err)
	assert.Equal(t, expectedJobName, job.Name)
	assert.Equal(t, aiplatformpb.JobState_JOB_STATE_PENDING, job.State)

	mockClient.AssertExpectations(t)

	// Verify request details
	call := mockClient.Calls[0]
	req := call.Arguments.Get(1).(*aiplatformpb.CreateBatchPredictionJobRequest)

	assert.Equal(t, "projects/my-project/locations/us-central1", req.Parent)
	assert.Equal(t, "test-job", req.BatchPredictionJob.DisplayName)
	assert.Equal(t, "gemini-2.5-flash", req.BatchPredictionJob.Model)
	assert.Equal(t, "jsonl", req.BatchPredictionJob.InputConfig.InstancesFormat)
	assert.Equal(t, "gs://bucket/input.jsonl", req.BatchPredictionJob.InputConfig.GetGcsSource().Uris[0])
	assert.Equal(t, "jsonl", req.BatchPredictionJob.OutputConfig.PredictionsFormat)
	assert.Equal(t, "gs://bucket/output/", req.BatchPredictionJob.OutputConfig.GetGcsDestination().OutputUriPrefix)
}
