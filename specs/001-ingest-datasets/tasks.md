# Tasks: Ingest Prompt Injection Datasets

**Feature**: Ingest Prompt Injection Datasets (`001-ingest-datasets`)
**Status**: In Progress

## Dependencies

- **Phase 1 (Setup)**: No dependencies.
- **Phase 2 (Foundational)**: Depends on Phase 1.
- **Phase 3 (User Story 1)**: Depends on Phase 2.
- **Phase 4 (User Story 2)**: Depends on Phase 3.
- **Phase 5 (Polish)**: Depends on Phase 3.

## Implementation Strategy

We will evolve the existing PoC in `cmd/dataset-loader/main.go` into a production-ready Service using Hexagonal Architecture.
- **MVP (Phase 3)**: A working CLI command that downloads one dataset and upserts to Pinecone.
- **Polish (Phase 5)**: Robust error handling, logging, and rate limit management.

## Phase 1: Setup

**Goal**: Initialize the project structure and clean up the existing PoC.

- [x] T001 Define configuration struct in `internal/app/dataset_loader/config.go` (porting from main.go)
- [x] T002 Initialize `internal/app/dataset_loader/service.go` with an empty Service struct
- [x] T003 Create `cmd/dataset-loader/wire.go` for dependency injection setup
- [x] T004 Refactor `cmd/dataset-loader/main.go` to use Wire and the new Service (removing inline PoC logic)

## Phase 2: Foundational

**Goal**: Implement the platform adapters required for the core logic.

- [ ] T005 [P] Create `internal/platform/huggingface/client.go` to handle Parquet file downloading and reading
- [ ] T006 [P] Create `internal/platform/pinecone/client.go` with `UpsertBatch` method (using Pinecone Go SDK v4)
- [ ] T007 [P] Create `internal/platform/pinecone/interface.go` defining the `VectorStore` interface

## Phase 3: User Story 1 - Ingest Standard Injection Datasets (P1)

**Goal**: As a System Administrator, I want to trigger an ingestion job that pulls known prompt injection datasets from HuggingFace and stores them in Pinecone.

**Independent Test**: Run the command with `HF_DATASET_ID` set to a small dataset (or subset) and verify vectors appear in Pinecone.

- [x] T008 [US1] Define domain entities in `internal/app/dataset_loader/domain.go` (`SourceRecord`, `IngestionRecord`)
- [x] T009 [US1] Implement mapping logic (Parquet Row -> Ingestion Record) in `internal/app/dataset_loader/mapper.go`
- [x] T010 [US1] Implement `Service.Ingest` method in `internal/app/dataset_loader/service.go` orchestrating HF download and Pinecone upsert
- [x] T011 [US1] Update `cmd/dataset-loader/wire.go` to inject HF and Pinecone clients into the Service
- [x] T012 [US1] Manual verification: Run `go run cmd/dataset-loader/main.go` and check Pinecone console

## Phase 4: User Story 2 - Idempotent Re-ingestion (P2)

**Goal**: As an Administrator, I want the ingestion process to be idempotent to prevent duplicate data.

**Independent Test**: Run the ingestion job twice and verify vector count remains constant.

- [ ] T013 [US2] Ensure ID generation in `mapper.go` uses deterministic SHA256 hash of the content
- [ ] T014 [US2] Manual verification: Run the ingestion command twice and verify logs/Pinecone stats

## Phase 5: Polish & Cross-Cutting Concerns

**Goal**: Production readiness (logging, error handling).

- [ ] T015 [P] Add structured logging (slog) throughout the `Service` and adapters
- [ ] T016 [P] Implement batching in `huggingface/client.go` or `service.go` to handle memory usage for large files
- [ ] T017 [P] Add graceful error handling for network timeouts (retries)
