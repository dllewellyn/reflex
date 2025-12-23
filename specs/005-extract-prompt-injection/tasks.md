# Tasks: Extract Prompt Injection Strings

**Feature**: `005-extract-prompt-injection`
**Status**: Draft

## Phase 1: Setup
*Goal: Initialize project structure and new configuration.*

- [x] T001 Create `cmd/extract-injections` directory and `main.go` scaffold in `cmd/extract-injections/main.go`
- [x] T002 Create `internal/app/extract` directory for application logic in `internal/app/extract/service.go`
- [x] T003 Create `internal/platform/genai` directory for Vertex AI client in `internal/platform/genai/client.go`
- [x] T004 Define configuration struct (including GCS, Pinecone, Vertex env vars) in `internal/app/extract/config.go`
- [x] T005 Create prompt template file in `prompts/extract-injection.prompt.yml`

## Phase 2: Foundational Components
*Goal: Implement platform clients and core interfaces required by all stories.*

- [x] T006 [P] Implement Vertex AI GenAI client wrapper for `GenerateContent` in `internal/platform/genai/client.go`
- [x] T007 [P] Define `GenAIClient` interface in `internal/platform/genai/interface.go` and generate mocks in `internal/platform/genai/mock_client.go`
- [x] T008 [P] Ensure Pinecone client in `internal/platform/pinecone/client.go` supports necessary upsert operations (already exists, verify interface)
- [x] T009 [P] Ensure GCS client in `internal/platform/gcs/client.go` supports reading/listing files (already exists)

## Phase 3: User Story 1 - Extract and Upsert Known Attacks (Priority: P1)
*Goal: Process batch results, extract injections using LLM, and upsert to Pinecone.*
*Independent Test*: Run command on mock batch file, verify LLM calls for positive cases and Pinecone upserts.

- [x] T010 [US1] Create `BatchResult` struct to parse Vertex AI Batch output JSONL in `internal/app/extract/domain.go`
- [x] T011 [US1] Implement `BatchReader` to read and parse JSONL files from GCS in `internal/app/extract/reader.go`
- [x] T012 [US1] Implement `Extractor` logic to call GenAI client with prompt and transcript in `internal/app/extract/extractor.go`
- [x] T013 [US1] Implement `Processor` to orchestrate GCS read -> Filter (is_injection) -> Extract -> Upsert flow in `internal/app/extract/processor.go`
- [x] T014 [US1] Wire up `Service` in `internal/app/extract/service.go` to use `Processor`
- [x] T015 [US1] Implement `main.go` with `wire` dependency injection and CLI flag parsing in `cmd/extract-injections/main.go`
- [x] T016 [US1] Create wire generation file in `cmd/extract-injections/wire.go` and run `wire`
- [ ] T017 [US1] Implement integration test with mock GCS/Vertex/Pinecone in `features/extract_e2e_test.go`

## Phase 4: User Story 2 - Idempotency and Tracking (Priority: P2)
*Goal: Prevent duplicate processing and duplicate vector entries.*
*Independent Test*: Run command twice, verify second run skips or deduplicates.

- [x] T018 [US2] Update `Processor` to generate deterministic IDs for Pinecone vectors (SHA256 of payload) in `internal/app/extract/processor.go`
- [ ] T019 [US2] Implement logic to check if interaction ID was already processed (optional, based on requirement "skip already-processed") or rely on idempotent upserts. *Decision: Rely on idempotent upserts for P2 MVP as per plan research.*
- [x] T020 [US2] Add unit test for deterministic ID generation in `internal/app/extract/processor_test.go`
- [x] T021 [US2] Verify idempotent behavior in `features/extract_e2e_test.go`

## Phase 5: Polish & Cross-Cutting
*Goal: Logging, observability, and final cleanup.*

- [x] T022 Add OpenTelemetry tracing to `Extract` and `Upsert` operations in `internal/app/extract/processor.go`
- [x] T023 Add structured logging for extraction results (found vs skipped) in `internal/app/extract/processor.go`
- [x] T024 Implement dry-run mode logic in `internal/app/extract/processor.go` and `cmd/extract-injections/main.go`
- [x] T025 Run full build and lint check `make build` and `golangci-lint run`

## Dependencies

- Phase 1 & 2 must be complete before Phase 3.
- Phase 3 (US1) must be complete before Phase 4 (US2) logic is fully verified, though T018 can be done in parallel.

## Implementation Strategy

1.  **MVP**: Implement Phase 1, 2, and 3 (US1). This delivers the core value of extracting and saving injections.
2.  **Robustness**: Implement Phase 4 (US2) to ensure the system is safe to re-run.
3.  **Production Ready**: Complete Phase 5 for observability and operational safety (dry-run).
