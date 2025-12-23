# Implementation Plan: Ingest Prompt Injection Datasets

**Branch**: `001-ingest-datasets` | **Date**: 2025-12-15 | **Spec**: [specs/001-ingest-datasets/spec.md](specs/001-ingest-datasets/spec.md)
**Input**: Feature specification from `specs/001-ingest-datasets/spec.md`

## Summary

Implement a Go-based ingestion job that fetches prompt injection datasets from HuggingFace, processes them into vector embeddings, and stores them in a Pinecone vector index. This establishes the baseline threat knowledge base for the system.

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: 
- `github.com/pinecone-io/go-pinecone` (Pinecone Client)
- `github.com/google/wire` (Dependency Injection)
- `go.opentelemetry.io/otel` (Observability)
- `github.com/spf13/cobra` (CLI) or standard `flag` (NEEDS CLARIFICATION on CLI preference)
**Storage**: Pinecone (Vector Database), Memory (for buffering batches)
**Testing**: Go standard testing
**Target Platform**: Linux / Docker (Cloud Run Job or Kubernetes Job)
**Project Type**: CLI / Batch Job
**Performance Goals**: Ingest 10k records < 15 mins.
**Constraints**: Handle API rate limits (HF and Pinecone).
**Scale/Scope**: ~10-100k initial records.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **BDD Tests:** All features must have BDD test cases defined in this plan.
- [x] **Observability:** Logging and OpenTelemetry integration must be planned.
- [x] **E2E Testing:** Plan includes testing against real infrastructure (Pinecone) in CI.
- [x] **Infrastructure:** Infrastructure requirements identified? (Pinecone Index)
- [x] **Serverless:** All new infrastructure is serverless/scale-to-zero? (Cloud Run Job)
- [x] **Secrets:** Secrets management strategy defined? (`PINECONE_API_KEY`, `HF_TOKEN`)
- [x] **Schemas:** JSON Schemas/OpenAPI defined for external interfaces? (Dataset Record Schema)
- [x] **Codegen:** Plan includes automated code generation from schemas?

## BDD Test Cases

| Feature | Scenario | Given | When | Then |
| --- | --- | --- | --- | --- |
| Ingestion | Ingest Standard Dataset | A valid HuggingFace dataset config and empty Pinecone index | The ingestion job is triggered | Vectors are stored in Pinecone with correct metadata and count matches source |
| Ingestion | Idempotency | A dataset has already been ingested | The ingestion job is triggered again | No duplicate vectors are created, count remains constant |
| Ingestion | Rate Limit Handling | A mocked rate limit response from HF API | The ingestion job is triggered | The system retries with backoff and eventually succeeds |
| Ingestion | Invalid Config | An invalid dataset ID | The ingestion job is triggered | The system exits with a specific error code and logs the failure |

## Project Structure

### Documentation (this feature)

```text
specs/001-ingest-datasets/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── config_schema.yaml # Configuration schema for the job
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
cmd/
└── dataset-loader/
    ├── main.go          # Entry point
    └── wire.go          # Dependency injection wiring

internal/
├── app/
│   └── dataset_loader/
│       ├── service.go   # Core business logic (orchestration)
│       └── service_test.go
└── platform/
    ├── pinecone/
    │   ├── client.go    # Pinecone SDK wrapper
    │   └── client_test.go
    └── huggingface/
        ├── client.go    # HF API client
        └── client_test.go
```

**Structure Decision**: Standard Go CLI structure with hexagonal/clean architecture principles (app core, platform adapters).