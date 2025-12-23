package schema

import "time"

// Interaction represents a single turn in a conversation (User Input + Model Output).
type Interaction struct {
	InteractionID  string      `json:"interaction_id"`
	ConversationID string      `json:"conversation_id"`
	Timestamp      time.Time   `json:"timestamp"`
	UserInput      UserInput   `json:"user_input"`
	ModelOutput    ModelOutput `json:"model_output"`
}

// UserInput represents the user's prompt.
type UserInput struct {
	Content  string            `json:"content"`
	Metadata UserInputMetadata `json:"metadata"`
}

// UserInputMetadata contains context about the user's input.
type UserInputMetadata struct {
	UserID    string `json:"user_id"`
	SourceIP  string `json:"source_ip"`
	UserAgent string `json:"user_agent"`
}

// ModelOutput represents the LLM's response.
type ModelOutput struct {
	Content  string              `json:"content"`
	Metadata ModelOutputMetadata `json:"metadata"`
}

// ModelOutputMetadata contains context about the model's output.
type ModelOutputMetadata struct {
	ModelID      string `json:"model_id"`
	LatencyMS    int64  `json:"latency_ms"`
	FinishReason string `json:"finish_reason"`
}
