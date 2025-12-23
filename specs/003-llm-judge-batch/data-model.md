# Data Model

## Entities

### InteractionEvent
*Source of Truth: `specifications/schemas/interaction-event.schema.json`*

Represents a single raw interaction (turn) in a conversation.

| Field | Type | Description |
|---|---|---|
| `interaction_id` | UUID | Unique ID for the message. |
| `conversation_id` | UUID | Grouping ID for the session. |
| `timestamp` | Timestamp | When the event occurred. |
| `role` | Enum(user, model) | Who generated the content. |
| `content` | String | The text content. |

### VertexBatchInput
*Source of Truth: `specifications/schemas/vertex-batch-input.schema.json`*

The format required by Vertex AI Batch Prediction.

| Field | Type | Description |
|---|---|---|
| `request` | Object | Container for the prompt. |
| `request.contents` | Array | List of content parts. |

### SecurityAlert
*Source of Truth: `specifications/schemas/security-alert.schema.json`*

The output event produced when a threat is detected.

| Field | Type | Description |
|---|---|---|
| `alert_id` | UUID | Unique ID for the alert. |
| `severity` | Enum | Threat level. |

### SecurityJudgePrompt (NEW)
*Format: GitHub Models Prompt (`.prompt.yml`)*

Location: `prompts/security-judge.prompt.yml`

```yaml
name: Security Judge
description: Analyzes conversation transcripts for prompt injection.
model: gemini-2.5-flash
messages:
  - role: system
    content: "You are a security AI..."
  - role: user
    content: "Analyze this transcript:\n\n{{conversation_transcript}}"
```

## Data Stores
*   **GCS**: Raw logs, Batch Input/Output.
*   **Kafka**: Security Alerts.
*   **Local FS**: `prompts/*.prompt.yml` (Source of Truth for LLM instructions).
