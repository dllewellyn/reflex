package extract

import "time"

// BatchResult represents a single line in the Vertex AI Batch Prediction output file.
type BatchResult struct {
	EventID  string   `json:"event_id"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Commit   func()   `json:"-"`
}

type Response struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type Request struct {
	Contents []Content `json:"contents"`
}

// JudgeOutput represents the JSON string inside the prediction text.
type JudgeOutput struct {
	IsPromptInjection bool    `json:"is_prompt_injection"`
	Confidence        float64 `json:"confidence"`
	Severity          string  `json:"severity"`
	Analysis          string  `json:"analysis"`
}

// ExtractedAttack represents a successful extraction.
type ExtractedAttack struct {
	OriginalInteractionID string
	InjectionPayload      string
	ExtractedAt           time.Time
}
