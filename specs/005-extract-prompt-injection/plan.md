# Implementation Plan - Extract Prompt Injection Strings

**Feature**: `005-extract-prompt-injection`
**Status**: Draft

## Technical Context

The system currently uses a **Daily Batch Job** to detect prompt injections in conversation logs stored in GCS. The results are output to GCS in JSONL format. This feature adds a **Feedback Loop** by processing those results, extracting the specific injection strings using an LLM, and updating the **Pinecone Vector Database** to improve future detection capabilities.

### New Components
1.  **Extraction CLI (`cmd/extract-injections`)**: A standalone Go binary.
2.  **Extraction Prompt**: A new prompt template for the LLM.

4.  **GenAI Client**: A client to interact with Vertex AI's Generative AI API (online prediction) for extraction.

### Architecture Updates
*   **New Flow**: `GCS (Batch Results)` -> `Extract CLI` -> `Vertex AI (GenAI)` -> `Vertex AI (Embed)` -> `Pinecone`.

### Technology Stack
*   **Language**: Go 1.23+
*   **APIs**:
    *   `cloud.google.com/go/vertexai/genai` (for LLM extraction)
    *   `cloud.google.com/go/aiplatform/apiv1` (for Embeddings, using standard client or REST wrapper if needed)
    *   `github.com/pinecone-io/go-pinecone/v4` (for Vector Storage)
    *   `cloud.google.com/go/storage` (for reading Batch Results)

## Constitution Check

| Principle | Compliance Check |
|---|---|
| **I. BDD Test Cases** | **Pending**: Scenarios defined in Spec, will be formalized in `features/` |
| **II. Directory Structure** | **Compliant**: Follows `cmd/`, `internal/app/`, `internal/platform/` structure. |
| **III. Core Architecture** | **Compliant**: Extends the Kafka/GCS/Vertex architecture. |
| **IV. Primary Language** | **Compliant**: Go. |
| **V. Interfaces** | **Compliant**: Will define interfaces for new Embedding and GenAI clients. |
| **VI. Dependency Injection** | **Compliant**: Will use `wire`. |
| **VII. Clean Code** | **Compliant**. |
| **VIII. Secrets** | **Compliant**: Uses `.env` and standard auth. |
| **IX. Observability** | **Compliant**: Will verify OpenTelemetry integration. |
| **X. Continuous Verification** | **Compliant**: Will run build/tests. |
| **XI. E2E Testing** | **Compliant**: Will add `features/extract_e2e_test.go`. |
| **XII. Serverless** | **Compliant**: Runs as a scheduled job (Cloud Run Job). |
| **XIII. Infrastructure** | **Compliant**: Uses existing GCS, Vertex, Pinecone. |
| **XIV. Schema-Driven** | **Compliant**: Uses Schemas for IO. |

## Phase 0: Outline & Research

*Status: Completed*

- [x] **Research Embedding Generation**: Decided to use Vertex AI Embeddings API to ensure compliance with "System MUST generate vectors" requirement.
- [x] **Research LLM Extraction**: Decided to use Vertex AI `genai` package for online extraction of positive results.
- [x] **Research Idempotency**: Decided on `SHA256(payload)` as deterministic ID for Pinecone to handle deduplication naturally.

## Phase 1: Design & Contracts

*Status: Completed*

- [x] **Data Model**: Defined `BatchResult`, `ExtractionPrompt`, `VectorEntry`.
- [x] **Quickstart**: Created `quickstart.md`.
- [x] **Agent Context**: Updated.

## Phase 2: Core Implementation

### 1. Platform Layers

-   **GenAI Client**: Implement `internal/platform/genai/client.go` for text generation.

### 2. Application Logic (`internal/app/extract`)
-   **Service**: Orchestrates the flow.
-   **Processor**: Handles the `GCS -> Filter -> Extract -> Embed -> Upsert` pipeline for a batch.

### 3. CLI (`cmd/extract-injections`)
-   **Main**: Wires everything up using `wire`.
-   **Config**: Env var processing.

### 4. Prompt
-   Create `prompts/extract-injection.prompt.yml`.

## Phase 3: Testing & Verification

### 1. Unit Tests
-   Mock `EmbeddingClient` and `GenAIClient`.
-   Test `Processor` logic with mock batch results.

### 2. Integration/E2E
-   **Feature File**: `features/extract.feature`.
-   **Test**: `features/extract_e2e_test.go` using real (or emulator) GCS and mocked Vertex/Pinecone if possible, or real dev environment.

### 3. Verification
-   Run `go test ./...`.
-   Run `make build`.