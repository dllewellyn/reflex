package batch

import (
	"fmt"
	"os"

	"strings"

	"gopkg.in/yaml.v3"
)

// Prompt represents the structure of the GitHub Models prompt file.
type Prompt struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Model       string    `yaml:"model"`
	Messages    []Message `yaml:"messages"`
}

// Message represents a single message in the prompt conversation.
type Message struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

// LoadPrompt reads and parses a YAML prompt file.
func LoadPrompt(path string) (*Prompt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file: %w", err)
	}

	var prompt Prompt
	if err := yaml.Unmarshal(data, &prompt); err != nil {
		return nil, fmt.Errorf("failed to parse prompt yaml: %w", err)
	}

	return &prompt, nil
}

// FormatPrompt substitutes placeholders in the prompt messages.
func FormatPrompt(messages []Message, transcript string) []Message {
	formatted := make([]Message, len(messages))
	for i, msg := range messages {
		formatted[i] = Message{
			Role:    msg.Role,
			Content: strings.ReplaceAll(msg.Content, "{{conversation_transcript}}", transcript),
		}
	}
	return formatted
}
