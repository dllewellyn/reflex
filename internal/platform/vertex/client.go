package vertex

import (
	"context"

	"cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/googleapis/gax-go/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Client is a wrapper around the Vertex AI JobClient.
type Client struct {
	client *aiplatform.JobClient
}

// NewClient creates a new Vertex AI JobClient wrapper.
func NewClient(c *aiplatform.JobClient) *Client {
	return &Client{client: c}
}

// CreateBatchPredictionJob creates a batch prediction job in Vertex AI.
func (c *Client) CreateBatchPredictionJob(ctx context.Context, req *aiplatformpb.CreateBatchPredictionJobRequest, opts ...gax.CallOption) (*aiplatformpb.BatchPredictionJob, error) {
	tr := otel.Tracer("vertex-client")
	ctx, span := tr.Start(ctx, "CreateBatchPredictionJob")
	defer span.End()

	span.SetAttributes(
		attribute.String("vertex.display_name", req.BatchPredictionJob.DisplayName),
		attribute.String("vertex.model", req.BatchPredictionJob.Model),
	)

	return c.client.CreateBatchPredictionJob(ctx, req, opts...)
}
