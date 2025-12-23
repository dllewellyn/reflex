# Research & Technical Decisions

## Decisions

### 1. Data Source Format
*   **Decision**: The Batch Job will consume data from Google Cloud Storage (GCS) in **JSON Lines (JSONL)** format.
*   **Rationale**: The existing `loader` service (`internal/app/loader/service.go`) writes data in this format. JSONL is efficient for batch processing as it allows line-by-line reading without parsing a massive JSON array.

### 2. GCS Directory Structure
*   **Decision**: The Batch Job will scan paths matching the pattern: `raw/*/<YYYY>/<MM>/<DD>/*/*.jsonl`.
*   **Rationale**: Aligns with existing loader output.

### 3. Data Schema
*   **Decision**: Map internal data to `InteractionEvent`.
*   **Schema**: Defined in `contracts/interaction-event.schema.json`.

### 4. Vertex AI Integration
*   **Decision**: Use Vertex AI Batch Prediction API via `aiplatform` library.
*   **Model**: `gemini-2.5-flash`.
*   **Trigger**: `CreateBatchPredictionJob`.

### 5. Prompt Management (NEW)
*   **Decision**: Store the Judge Prompt in **GitHub Models format** (`.prompt.yml`).
*   **Rationale**: Standardization and compatibility with GitHub's ecosystem.
*   **File Path**: `prompts/security-judge.prompt.yml`.
*   **Usage**: The Batch Job will read this file at runtime (or build time) to construct the `contents` for the Vertex Batch Input. It will substitute the `{{conversation_transcript}}` placeholder with the actual aggregated logs.