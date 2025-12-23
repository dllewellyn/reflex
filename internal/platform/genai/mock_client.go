package genai

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of ClientInterface.
type MockClient struct {
	mock.Mock
}

// Ensure MockClient implements ClientInterface
var _ ClientInterface = (*MockClient)(nil)

func (m *MockClient) GenerateContent(ctx context.Context, modelName, prompt string) (string, error) {
	args := m.Called(ctx, modelName, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
