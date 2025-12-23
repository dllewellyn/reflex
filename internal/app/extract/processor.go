package extract

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dllewellyn/reflex/internal/platform/pinecone"
	"go.opentelemetry.io/otel"
)

type Processor struct {
	reader    ResultReader
	extractor *Extractor
	pinecone  pinecone.VectorStore
	config    Config
}

func NewProcessor(reader ResultReader, extractor *Extractor, pc pinecone.VectorStore, cfg Config) *Processor {
	return &Processor{
		reader:    reader,
		extractor: extractor,
		pinecone:  pc,
		config:    cfg,
	}
}

func (p *Processor) Process(ctx context.Context) error {
	tr := otel.Tracer("extract-processor")
	ctx, span := tr.Start(ctx, "Processor.Process")
	defer span.End()

	results, errCh, teardown := p.reader.ReadResults(ctx)
	defer teardown()

	batchSize := 96
	var batch []*pinecone.InputRecord
	var commits []func()

	for result := range results {
		// Check for error from reader
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		default:
		}

		records, err := p.processResult(ctx, result)
		if err != nil {
			slog.Error("Failed to process result", "event_id", result.EventID, "error", err)
			continue
		}

		// Always collect the commit function, regardless of whether we produce records
		// If processResult returns error, we might log and continue, effectively skipping it
		// but we still need to commit it so we don't get stuck.
		if result.Commit != nil {
			commits = append(commits, result.Commit)
		}

		if len(records) == 0 {
			continue
		}

		if p.config.DryRun {
			for _, record := range records {
				slog.Info("Dry Run: Would upsert", "event_id", result.EventID, "id", record.ID, "text_len", len(record.Text))
			}
			continue
		}

		batch = append(batch, records...)

		if len(batch) >= batchSize {
			if err := p.upsertBatch(ctx, batch); err != nil {
				return err
			}
			batch = batch[:0]
			for _, commit := range commits {
				commit()
			}
			commits = commits[:0]
		}
	}

	// Check error after loop
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	default:
	}

	slog.Info("Upserting remaining batch", "count", len(batch))

	if len(batch) > 0 {
		if err := p.upsertBatch(ctx, batch); err != nil {
			return err
		}
		for _, commit := range commits {
			commit()
		}
	} else if len(commits) > 0 {
		// Flush remaining commits if any (e.g. from skipped items)
		for _, commit := range commits {
			commit()
		}
	}

	return nil
}

func (p *Processor) processResult(ctx context.Context, result BatchResult) ([]*pinecone.InputRecord, error) {
	if len(result.Response.Candidates) == 0 || len(result.Response.Candidates[0].Content.Parts) == 0 {
		slog.Warn("Received message without any candidates", "event_id", result.EventID)
		return nil, nil
	}

	slog.Info("Processing event", "event_id", result.EventID)

	judgeOutput, err := p.parseJudgeOutput(result.Response.Candidates[0].Content.Parts[0].Text)
	if err != nil {
		// Original behavior: slog.Warn and continue
		slog.Warn("Failed to parse judge output", "event_id", result.EventID, "error", err)
		return nil, nil // Return nil, nil to indicate skip
	}

	slog.Info("Judge analysis complete",
		"event_id", result.EventID,
		"is_prompt_injection", judgeOutput.IsPromptInjection,
		"confidence", judgeOutput.Confidence,
		"severity", judgeOutput.Severity,
	)

	if !judgeOutput.IsPromptInjection {
		slog.Debug("Not a prompt injection - skipping", "event_id", result.EventID, "confidence", judgeOutput.Confidence)
		return nil, nil
	}

	if len(result.Request.Contents) == 0 || len(result.Request.Contents[0].Parts) == 0 {
		slog.Error("No request contents")
		return nil, nil
	}

	// Extract
	transcript := result.Request.Contents[0].Parts[0].Text
	injections, err := p.extractor.Extract(ctx, transcript)
	if err != nil {
		slog.Error("Failed to extract injection", "event_id", result.EventID, "error", err)
		return nil, nil
	}

	slog.Info("Extraction complete", "event_id", result.EventID, "candidates_count", len(injections))

	var records []*pinecone.InputRecord
	for _, injection := range injections {
		if injection == "" {
			slog.Error("Empty injection")
			continue
		}

		id := generateID(injection)
		record := &pinecone.InputRecord{
			ID:   id,
			Text: injection,
			Metadata: map[string]interface{}{
				"source":       "auto-extracted",
				"label":        "injection",
				"extracted_at": time.Now().Format(time.RFC3339),
			},
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("expected prompt injections to be extracted but instead got 0 records: %s", transcript)
	}

	return records, nil
}

func (p *Processor) parseJudgeOutput(rawOutput string) (JudgeOutput, error) {
	rawOutput = strings.TrimPrefix(rawOutput, "```json")
	rawOutput = strings.TrimSuffix(rawOutput, "```")
	rawOutput = strings.TrimSpace(rawOutput)

	var judgeOutput JudgeOutput
	if err := json.Unmarshal([]byte(rawOutput), &judgeOutput); err != nil {
		return JudgeOutput{}, fmt.Errorf("failed to parse judge output: %w raw: %s", err, rawOutput)
	}
	return judgeOutput, nil
}

func (p *Processor) upsertBatch(ctx context.Context, batch []*pinecone.InputRecord) error {
	if p.config.DryRun {
		return nil
	}
	if err := p.pinecone.UpsertInputs(ctx, batch); err != nil {
		return fmt.Errorf("failed to upsert batch: %w", err)
	}
	slog.Info("Upserted batch", "count", len(batch))
	return nil
}

func generateID(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
