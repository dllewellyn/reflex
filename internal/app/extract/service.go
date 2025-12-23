package extract

import (
	"context"
	"log/slog"
)

// Service implements the extraction service.
type Service struct {
	processor *Processor
}

// NewService creates a new instance of the extraction service.
func NewService(processor *Processor) *Service {
	return &Service{
		processor: processor,
	}
}

// Run executes the extraction process.
func (s *Service) Run(ctx context.Context) error {
	slog.Info("Starting extraction service")
	return s.processor.Process(ctx)
}
