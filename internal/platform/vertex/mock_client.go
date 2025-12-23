package vertex

import (
	"context"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/mock"
)

// MockJobClient is a mock implementation of JobClient for testing.
type MockJobClient struct {
	mock.Mock
}

// CreateBatchPredictionJob mocks the CreateBatchPredictionJob method.
func (m *MockJobClient) CreateBatchPredictionJob(ctx context.Context, req *aiplatformpb.CreateBatchPredictionJobRequest, opts ...gax.CallOption) (*aiplatformpb.BatchPredictionJob, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aiplatformpb.BatchPredictionJob), args.Error(1)
}
