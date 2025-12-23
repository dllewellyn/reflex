package genai

import "context"

// ClientInterface defines the methods for interacting with the Generative AI model.
type ClientInterface interface {
	// GenerateContent generates text content from a prompt.
	GenerateContent(ctx context.Context, modelName, prompt string) (string, error)
	// Close closes the client connection.
	Close() error
}
