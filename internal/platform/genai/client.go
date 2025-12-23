package genai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// Client handles interaction with Generative AI.
type Client struct {
	client *genai.Client
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)

// NewClient creates a new Vertex AI GenAI client.
func NewClient(ctx context.Context, projectID, location string) (*Client, error) {
	cfg := &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	}

	c, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &Client{client: c}, nil
}

// GenerateContent generates text content from a prompt.
func (c *Client) GenerateContent(ctx context.Context, modelName, prompt string) (string, error) {
	resp, err := c.client.Models.GenerateContent(ctx, modelName, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", nil
	}

	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		result += part.Text
	}

	return result, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	// The generated client does not strictly require closing if sharing transport,
	// but we implement the interface.
	return nil
}
