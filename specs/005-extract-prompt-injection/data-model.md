# Data Model: Extract Prompt Injection

## Entities

### BatchResult
*Source: GCS (Output of `cmd/batch-job`)*

Represents a single prediction result from the Batch Prediction Job.

| Field | Type | Description |
|---|---|---|
| `prediction` | Object | The model's prediction output. |
| `prediction.candidates[0].content.parts[0].text` | JSON String | The raw JSON output from the Judge (e.g., `{"prompt_injection": true, ...}`). |
| `instance` | Object | The original input passed to the model (preserved by Vertex Batch). |
| `instance.request.contents[0].parts[0].text` | String | The conversation transcript analyzed. |

### ExtractionPrompt
*Format: GitHub Models Prompt (`.prompt.yml`)*

Location: `prompts/extract-injection.prompt.yml`

```yaml
name: Injection Extractor
description: Extracts the specific prompt injection payload from a conversation.
model: gemini-2.5-flash
messages:
  - role: system
    content: "You are a security analyst..."
  - role: user
    content: "Extract the injection string from:\n\n{{transcript}}"
```

### ExtractedAttack
*Internal Go Struct*

The result of the extraction process.

| Field | Type | Description |
|---|---|---|
| `original_interaction_id` | UUID | Link to the source interaction. |
| `injection_payload` | String | The exact string identified as an attack. |
| `confidence` | Float | (Optional) Model's confidence. |

### VectorEntry (Pinecone)
*Target: Pinecone Index*

| Field | Type | Description |
|---|---|---|
| `id` | String | Deterministic hash `SHA256(injection_payload)`. |
| `values` | []Float32 | Embedding vector (e.g., 768 dims). |
| `metadata` | Map | Key-value pairs. |
| `metadata.text` | String | The `injection_payload`. |
| `metadata.source` | String | "auto-extracted". |
| `metadata.original_id` | String | UUID of the source interaction. |
| `metadata.extracted_at` | Timestamp | ISO8601 time of extraction. |

```