# Implementation Tasks - Automate Batch Results Processing

**Feature Branch**: `006-process-batch-results`
**Feature Spec**: `specs/006-process-batch-results/spec.md`

## Implementation Strategy

- **MVP Scope**: Complete all User Stories (US1, US2, US3) as they form a single pipeline. US1/US2 create the producer, US3 updates the consumer.
- **Execution Order**:
    1.  **Setup**: Generate data structures.
    2.  **Producer (US1/US2)**: Build the Cloud Function to read GCS and write to Kafka.
    3.  **Consumer (US3)**: Update the downstream service to read from Kafka.
- **Testing**: Unit tests for parsers; Integration tests for Kafka connectivity; Manual/E2E verification for the full pipeline (upload -> function -> kafka -> consumer).

## Phase 1: Setup

**Goal**: Initialize project structures and generate code from schemas.

- [x] T001 Generate Go structs from `batch-result-event.schema.json` to `internal/platform/schema` using `make generate`

## Phase 2: Foundational

**Goal**: Establish core infrastructure definitions required for the Cloud Function.

- [x] T002 Define `google_cloudfunctions2_function` and IAM roles in `terraform/llm-judge-batch.tf`

## Phase 3: User Story 1 - Automatic Batch Ingestion

**Goal**: Enable automatic triggering of processing logic upon file upload.
**Story**: US1 (P1)
**Independent Test**: Deploy infrastructure and verify Cloud Function is triggered by GCS upload (check logs).

- [x] T003 [US1] Create Cloud Function main entrypoint in `cmd/batch-result-trigger/main.go`
- [x] T004 [US1] Implement GCS Event mapping and handler function in `cmd/batch-result-trigger/function.go`
- [x] T005 [US1] Unit test GCS event parsing in `cmd/batch-result-trigger/function_test.go`

## Phase 4: User Story 2 - Streaming Results to Kafka

**Goal**: Process the file content and stream to Kafka.
**Story**: US2 (P1)
**Independent Test**: Run the function locally with a sample file and verify messages in local Kafka/Redpanda.

- [x] T006 [P] [US2] Implement streaming JSONL reader in `internal/app/batch/stream_reader.go`
- [x] T007 [P] [US2] Implement Kafka producer for BatchResultEvent in `internal/app/batch/producer.go`
- [x] T008 [US2] Integrate Reader and Producer into Cloud Function handler in `cmd/batch-result-trigger/function.go`
- [x] T009 [US2] Add unit tests for streaming reader in `internal/app/batch/stream_reader_test.go`

## Phase 5: User Story 3 - Stream Consumer Update

**Goal**: Update the downstream service to consume from the new Kafka stream.
**Story**: US3 (P1)
**Independent Test**: Run `extract-injections` locally pointing to the Kafka topic and verify it processes messages.

- [ ] T010 [US3] Update `extract.Config` to include Kafka connection params in `internal/app/extract/config.go`
- [ ] T011 [US3] Implement Kafka consumer for BatchResultEvent in `internal/app/extract/consumer.go`
- [ ] T012 [US3] Update Service interface to accept stream source in `internal/app/extract/service.go`
- [ ] T013 [US3] Wire new consumer into application in `cmd/extract-injections/wire.go` and `cmd/extract-injections/main.go`
- [ ] T014 [US3] Remove legacy GCS polling logic from `internal/app/extract/extractor.go` (or relevant file)

## Phase 6: Polish & Cross-Cutting Concerns

**Goal**: Final verification and cleanup.

- [ ] T015 Verify End-to-End flow: Upload file -> Cloud Function -> Kafka -> Extract Service
- [ ] T016 Ensure all new components use structured logging (slog)

## Dependencies

1.  T001 (Schema) is required for T007 and T011.
2.  T002 (Infra) is required for deployment (T015).
3.  T003/T004 (Trigger) and T006/T007 (Logic) can be developed in parallel but must merge for T008.
4.  US3 (Consumer) depends on US2 (Producer) for meaningful E2E testing, but code can be written independently using the Schema.
