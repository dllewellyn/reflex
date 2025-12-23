# Research & Decisions: Extract Prompt Injection Strings

## Technical Context

The system already performs daily batch analysis using Vertex AI to detect prompt injections (`cmd/batch-job`). The results are stored in GCS. The goal of this feature is to parse these results, extract specific injection payloads from positive detections, and feed them back into the Pinecone vector database to improve future detection.

### Existing Components
- **Batch Job (`cmd/batch-job`)**: Triggers Vertex AI Batch Prediction.
- **Dataset Loader (`cmd/dataset-loader`)**: Ingests HF datasets into Pinecone. Shows precedent for batch processing and Pinecone upserts.
- **Pinecone Client (`internal/platform/pinecone`)**: Wrapper for Pinecone SDK v4. Currently supports `UpsertBatch`.
- **Vertex Client (`internal/platform/vertex`)**: Wrapper for Vertex AI Job Client.

### Unknowns & Clarifications

1.  **Embedding Generation**:
    *   **Context**: The `dataset-loader` uses `pinecone.Vector` which expects `Values []float32`. The `spec.md` FR-005 says "System MUST embed the extracted strings using the configured embedding model".
    *   **Finding**: The current `pinecone/client.go` requires `Values`. We must determine if we are using client-side embedding (e.g., via Vertex AI Embeddings API) or server-side Pinecone inference.
    *   **Decision**: We will leverage **Pinecone's integrated inference** to generate the embeddings. This means the `Upsert` operation will be provided with the extracted text, and Pinecone will handle its conversion to vectors. This simplifies our local implementation and relies on Pinecone's capabilities.
    *   **Action**: No separate embedding client is required on our side. We will ensure the Pinecone client is used to pass the text directly (e.g., in metadata or a specific field if supported by Pinecone's integrated inference API) and that the Pinecone index is configured to embed this text.

2.  **LLM Extraction Mechanism**:
    *   **Context**: We need to pass the conversation transcript to an LLM to extract the injection string.
    *   **Decision**: We will use the **Vertex AI GenerateContent API** (online prediction, not batch) for this processing step. Since the volume of *positive* injections should be relatively low compared to the total traffic, online processing per-positive-result is acceptable and simpler than chaining another batch job.
    *   **Prompt**: A new prompt template `prompts/extract-injection.prompt.yml` will be created.

3.  **Idempotency**:
    *   **Context**: We need to avoid re-processing the same batch result files or re-upserting the same injections.
    *   **Decision**:
        *   **Tracking**: We will use a "marker file" strategy in GCS (e.g., `gs://<bucket>/processed/<batch-job-id>.done`) or simply rely on the fact that Pinecone upserts are idempotent (same ID = overwrite).
        *   **ID Strategy**: The ID for the Pinecone vector should be deterministic based on the injection string itself (e.g., `SHA256(injection_string)`). This ensures that if the same attack is seen multiple times, it just updates the existing record (or is ignored), effectively handling deduplication automatically.

## Implementation Decisions

### 1. New Command: `cmd/extract-injections`
We will create a new CLI tool following the pattern of `cmd/batch-job`.
- **Inputs**: GCS URI pattern for batch results.
- **Logic**:
    1.  List and read GCS JSONL files.
    2.  Filter for `prompt_injection: true`.
    3.  Call Vertex AI (Gemini Flash) with extraction prompt.
    4.  Call Vertex AI (Embeddings) to vectorize extracted strings.
    5.  Upsert to Pinecone using deterministic IDs.

### 2. New Platform Components
-   **Embedding Client**: New interface and implementation for text embeddings.
-   **GenAI Client**: Logic to call `GenerateContent` (we might need to add this to `internal/platform/vertex` or a new package if the existing one is only for *Jobs*). *Correction*: The existing `internal/platform/vertex` is for *Job Service*. The `cmd/batch-job` imports `cloud.google.com/go/aiplatform/apiv1`. For online generation, we usually use `cloud.google.com/go/vertexai/genai`. We should add a new client wrapper for this.

### 3. Data Flow
`GCS (Batch Results)` -> `Reader` -> `Filter` -> `LLM (Extract)` -> `Embedding` -> `Pinecone`

## Rationale
-   **Vertex Embeddings**: Ensures consistency and control.
-   **Deterministic IDs**: Solves idempotency and deduplication elegantly.
-   **Separate CLI**: Decouples extraction from detection, allowing independent scheduling and scaling.
