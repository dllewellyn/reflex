package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TestPrompt represents a single entry in test_prompts.json
type TestPrompt struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PromptFile represents the partial structure of the prompt.yml file
// We will simply append text to the file to avoid complex YAML marshaling issues with comments, etc.
// but we need to know what to append.

func main() {
	// 1. Read test_prompts.json
	data, err := os.ReadFile("test_prompts.json")
	if err != nil {
		panic(err)
	}

	var prompts []TestPrompt
	if err := json.Unmarshal(data, &prompts); err != nil {
		panic(err)
	}

	// 2. Build the testData YAML block
	var sb strings.Builder
	sb.WriteString("\ntestData:\n")

	for _, p := range prompts {
		// Escape newlines and quotes in the input text for YAML safety
		safeText := strings.ReplaceAll(p.Text, "\"", "\\\"")
		safeText = strings.ReplaceAll(safeText, "\n", "\\n")

		sb.WriteString("  - conversation_transcript: \"" + safeText + "\"\n")

		// Map expected output based on type
		var expected bool
		if p.Type == "attack" {
			expected = true
		} else {
			expected = false
		}

		// We will construct the expected JSON string.
		// Since the prompt output contains other fields like confidence/analysis, exact match won't work perfectly
		// unless we mock those or use a different evaluator.
		// However, to fix the structure error, we must move evaluators out or use 'expected'.
		// Let's use 'expected' with a partial string for easier visual diff in UI, or for some evaluator to pick up.

		// For now, let's just put the core boolean expectation in 'expected'
		sb.WriteString(fmt.Sprintf("    expected: \"{\\\"is_prompt_injection\\\": %v}\"\n", expected))
	}

	// Add root-level evaluators
	sb.WriteString("\nevaluators:\n")
	sb.WriteString("  - name: Valid JSON Output\n")
	sb.WriteString("    string:\n")
	sb.WriteString("      contains: '\"is_prompt_injection\":'\n")
	sb.WriteString("  - name: Correctness (Similarity)\n")
	sb.WriteString("    uses: github/similarity\n")
	promptPath := "prompts/security-judge.prompt.yml"
	promptData, err := os.ReadFile(promptPath)
	if err != nil {
		panic(err)
	}

	promptContent := string(promptData)

	// Remove existing testData if present
	if idx := strings.Index(promptContent, "\ntestData:"); idx != -1 {
		promptContent = promptContent[:idx]
	}

	// 4. Append new testData
	newContent := promptContent + sb.String()

	// 5. Write back to file
	if err := os.WriteFile(promptPath, []byte(newContent), 0644); err != nil {
		panic(err)
	}

	fmt.Printf("Successfully injected %d test cases into %s\n", len(prompts), promptPath)
}
