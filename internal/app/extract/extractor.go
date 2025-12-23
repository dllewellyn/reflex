package extract

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/dllewellyn/reflex/internal/platform/genai"
	"go.opentelemetry.io/otel"
	"gopkg.in/yaml.v3"
)

type Extractor struct {
	client     genai.ClientInterface
	promptFile string
	template   *template.Template
	model      string
}

type PromptConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Model       string `yaml:"model"`
	Messages    []struct {
		Role    string `yaml:"role"`
		Content string `yaml:"content"`
	} `yaml:"messages"`
}

func NewExtractor(client genai.ClientInterface, promptFile string) *Extractor {
	return &Extractor{
		client:     client,
		promptFile: promptFile,
	}
}

// Initialize loads the prompt file and parses the template.
func (e *Extractor) Initialize() error {
	data, err := os.ReadFile(e.promptFile)
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %w", err)
	}

	var cfg PromptConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse prompt yaml: %w", err)
	}
	e.model = cfg.Model

	// Construct the full prompt from messages.
	var fullPromptBuilder strings.Builder
	for _, msg := range cfg.Messages {
		fullPromptBuilder.WriteString(msg.Content)
		fullPromptBuilder.WriteString("\n\n")
	}

	tmpl, err := template.New("prompt").Parse(fullPromptBuilder.String())
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}
	e.template = tmpl

	return nil
}

func (e *Extractor) Extract(ctx context.Context, transcript string) ([]string, error) {
	tr := otel.Tracer("extract-extractor")
	ctx, span := tr.Start(ctx, "Extractor.Extract")
	defer span.End()

	if e.template == nil {
		if err := e.Initialize(); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := e.template.Execute(&buf, map[string]string{"transcript": transcript}); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	resp, err := e.client.GenerateContent(ctx, e.model, buf.String())
	if err != nil {
		return nil, err
	}

	// Parse response (line separated, remove "None")
	lines := strings.Split(resp, "\n")
	var extractions []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.EqualFold(trimmed, "None") {
			continue
		}
		extractions = append(extractions, trimmed)
	}

	return extractions, nil
}
