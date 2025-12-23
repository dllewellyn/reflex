package vertex

import (
	"context"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/googleapis/gax-go/v2"
)

// JobClient defines the interface for interacting with Vertex AI Batch Prediction Jobs.
// It abstracts the underlying SDK client to facilitate testing.
type JobClient interface {
	CreateBatchPredictionJob(ctx context.Context, req *aiplatformpb.CreateBatchPredictionJobRequest, opts ...gax.CallOption) (*aiplatformpb.BatchPredictionJob, error)
}
