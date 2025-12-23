package ingestor

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dllewellyn/reflex/internal/app/ingestor/server"
	"github.com/dllewellyn/reflex/internal/platform/kafka"
	"github.com/dllewellyn/reflex/internal/platform/pinecone"
	"github.com/dllewellyn/reflex/internal/platform/schema"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Config struct {
	TopicName         string
	Port              string
}

type Service struct {
	producer      kafka.Producer
	vectorStore   pinecone.VectorStore
	topic         string
	port          string
}

// Ensure Service implements ServerInterface
var _ server.ServerInterface = (*Service)(nil)

func NewService(producer kafka.Producer, vectorStore pinecone.VectorStore, cfg Config) *Service {
	return &Service{
		producer:      producer,
		vectorStore:   vectorStore,
		topic:         cfg.TopicName,
		port:          cfg.Port,
	}
}

func (s *Service) Run(ctx context.Context) error {
	// Create the handler from the service implementation
	baseHandler := server.Handler(s)
	handler := otelhttp.NewHandler(baseHandler, "ingest")

	server := &http.Server{
		Addr:              ":" + s.port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("Starting ingestor HTTP server", "port", s.port, "topic", s.topic)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server shutdown failed", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// AnalyzeInteraction is the HTTP handler for analyzing interactions.
func (s *Service) AnalyzeInteraction(w http.ResponseWriter, r *http.Request) {
	var req server.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Failed to decode analyze request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Server-side timestamp
	timestamp := time.Now()

	// Transform to InteractionEvent
	event := schema.InteractionEvent{
		InteractionId:  req.InteractionId,
		ConversationId: req.ConversationId,
		Timestamp:      timestamp,
		Role:           schema.RoleUser, // Defaulting to user
		Content:        req.Prompt,
	}

	// Check for jailbreak attempts
	// Check for jailbreak attempts using sliding window chunking
	windowSize := 75
	overlap := 20
	chunks := chunkText(req.Prompt, windowSize, overlap)

	var maxScore float32
	var maxScoreMatchID string
	var isInjection bool

	for _, chunk := range chunks {
		matches, err := s.vectorStore.QueryInput(r.Context(), chunk, 1)
		if err != nil {
			slog.Error("Failed to query vector database", "error", err)
			// Continue checking other chunks or fail? Failsafe might be to log and continue,
			// but for security, if we can't check, we might want to be cautious.
			// However, the original code returned 500. Let's stick to that if any query fails for now,
			// or maybe just log error and continue if it's transient?
			// Given it's security, let's fail safe.
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if len(matches) > 0 {
			if matches[0].Score > maxScore {
				maxScore = matches[0].Score
				maxScoreMatchID = matches[0].ID
			}
		}
	}

	if maxScore > 0.84 {
		isInjection = true
		slog.Warn("Jailbreak attempt detected",
			"interaction_id", req.InteractionId,
			"score", maxScore,
			"matched_id", maxScoreMatchID)
	}

	ctx := r.Context()
	key := req.ConversationId

	if err := s.producer.Publish(ctx, s.topic, key, event); err != nil {
		slog.Error("Failed to publish to Kafka (main)", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := server.AnalyzeResponse{
		InteractionId:     req.InteractionId,
		Score:             maxScore,
		IsPromptInjection: isInjection,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

func chunkText(text string, windowSize, overlap int) []string {
	if text == "" {
		return []string{}
	}
	words := strings.Fields(text)
	if len(words) <= windowSize {
		return []string{text}
	}

	var chunks []string
	step := windowSize - overlap
	if step < 1 {
		step = 1
	}

	for i := 0; i < len(words); i += step {
		end := i + windowSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
		// If we've reached the end of the text, stop
		if end == len(words) {
			break
		}
	}
	return chunks
}
