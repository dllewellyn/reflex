package batch

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformer_CreateBatchRequest(t *testing.T) {
	tests := []struct {
		name           string
		promptMessages []Message
		transcript     string
		wantSystem     string
		wantContents   []Content
	}{
		{
			name: "single system message",
			promptMessages: []Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Analyze this: {{conversation_transcript}}"},
			},
			transcript: "hello world",
			wantSystem: "You are a helpful assistant.",
			wantContents: []Content{
				{Role: "user", Parts: []Part{{Text: "Analyze this: hello world"}}},
			},
		},
		{
			name: "no system message",
			promptMessages: []Message{
				{Role: "user", Content: "Just analyze: {{conversation_transcript}}"},
			},
			transcript: "foo bar",
			wantSystem: "",
			wantContents: []Content{
				{Role: "user", Parts: []Part{{Text: "Just analyze: foo bar"}}},
			},
		},
		{
			name: "multiple system messages",
			promptMessages: []Message{
				{Role: "system", Content: "System 1"},
				{Role: "system", Content: "System 2"},
				{Role: "user", Content: "User"},
			},
			transcript: "",
			wantSystem: "System 1\nSystem 2",
			wantContents: []Content{
				{Role: "user", Parts: []Part{{Text: "User"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewTransformer(&Prompt{Messages: tt.promptMessages})

			gotBytes, err := transformer.CreateBatchRequest(tt.transcript)
			require.NoError(t, err)

			var req BatchRequest
			err = json.Unmarshal(gotBytes, &req)
			require.NoError(t, err)

			// Verify System Instruction
			if tt.wantSystem == "" {
				assert.Nil(t, req.Request.SystemInstruction)
			} else {
				require.NotNil(t, req.Request.SystemInstruction)
				assert.Equal(t, tt.wantSystem, req.Request.SystemInstruction.Parts[0].Text)
			}

			// Verify Contents
			assert.Equal(t, tt.wantContents, req.Request.Contents)
		})
	}
}
