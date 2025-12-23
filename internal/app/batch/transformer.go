package batch

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Transformer handles the conversion of transcripts into Vertex AI Batch requests.
type Transformer struct {
	prompt *Prompt
}

// NewTransformer creates a new Transformer.
func NewTransformer(prompt *Prompt) *Transformer {
	return &Transformer{prompt: prompt}
}

// CreateBatchRequest converts a raw transcript into a JSONL line for the batch job.
func (t *Transformer) CreateBatchRequest(transcript string) ([]byte, error) {
	// Format prompt
	msgs := FormatPrompt(t.prompt.Messages, transcript)

	var contents []Content
	var systemParts []string

	for _, msg := range msgs {
		if msg.Role == "system" {
			systemParts = append(systemParts, msg.Content)
			continue
		}
		contents = append(contents, Content{
			Role:  msg.Role,
			Parts: []Part{{Text: msg.Content}},
		})
	}

	reqBody := BatchRequestBody{
		Contents: contents,
	}

	if len(systemParts) > 0 {
		reqBody.SystemInstruction = &SystemInstruction{
			Parts: []Part{{Text: strings.Join(systemParts, "\n")}},
		}
	}

	req := BatchRequest{
		Request: reqBody,
	}

	line, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	// Append newline as strictly required for JSONL
	return append(line, '\n'), nil
}

// Helper types for JSON marshaling
type BatchRequest struct {
	Request BatchRequestBody `json:"request"`
}

type BatchRequestBody struct {
	Contents          []Content          `json:"contents"`
	SystemInstruction *SystemInstruction `json:"system_instruction,omitempty"`
}

type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}
