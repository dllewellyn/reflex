package batch

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPrompt(t *testing.T) {
	// Create a temporary prompt file for testing
	content := `
name: Security Judge
description: Test Description
model: gemini-pro
messages:
  - role: system
    content: "System Prompt"
  - role: user
    content: "User Prompt {{conversation_transcript}}"
`
	tmpfile, err := os.CreateTemp("", "prompt_*.yaml")
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Logf("Failed to remove temporary file: %v", err)
		}
	}()

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	// Test loading the prompt
	prompt, err := LoadPrompt(tmpfile.Name())
	require.NoError(t, err)

	assert.Equal(t, "Security Judge", prompt.Name)
	assert.Equal(t, "Test Description", prompt.Description)
	assert.Equal(t, "gemini-pro", prompt.Model)
	require.Len(t, prompt.Messages, 2)
	assert.Equal(t, "system", prompt.Messages[0].Role)
	assert.Equal(t, "System Prompt", prompt.Messages[0].Content)
	assert.Equal(t, "user", prompt.Messages[1].Role)
	assert.Equal(t, "User Prompt {{conversation_transcript}}", prompt.Messages[1].Content)
}

func TestFormatPrompt(t *testing.T) {
	messages := []Message{
		{Role: "system", Content: "System"},
		{Role: "user", Content: "Analyze: {{conversation_transcript}}"},
	}

	formatted := FormatPrompt(messages, "My Conversation")
	require.Len(t, formatted, 2)
	assert.Equal(t, "System", formatted[0].Content)
	assert.Equal(t, "Analyze: My Conversation", formatted[1].Content)
}
