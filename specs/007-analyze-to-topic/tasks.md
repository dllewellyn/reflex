# Tasks: Analyze to Topic

**Input**: Design documents from `/specs/007-analyze-to-topic/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: The examples below include test tasks. Tests are a required part of the development workflow.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 [P] Add `ANALYSIS_TOPIC` variable to `terraform/variables.tf` and `terraform/ingestor.tf`
- [ ] T002 [P] Update `specifications/ingestor-api.yaml` with `/analyze` endpoint definition (ref `contracts/analyze.yaml`)
- [ ] T003 Generate Go code from OpenAPI spec (`make generate-api`)
- [ ] T003a Verify project builds successfully (`go build ./cmd/ingestor/...`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [ ] T004 Update `cmd/ingestor/main.go` `Config` struct to read `KAFKA_ANALYSIS_TOPIC` from env
- [ ] T005 Update `IngestorConfig` in `cmd/ingestor/main.go` and `internal/app/ingestor/service.go` to include `AnalysisTopicName`
- [ ] T005a Verify project builds successfully (`go build ./cmd/ingestor/...`)

---

## Phase 3: User Story 1 - Real-time Analysis (Priority: P1) 🎯 MVP

**Goal**: Enable clients to submit interaction data for asynchronous security analysis.

**Independent Test**: `PostAnalyze` unit test and integration test verifying Kafka publication.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T006 [P] [US1] Create unit test for `PostAnalyze` in `internal/app/ingestor/service_test.go` (verifying 202 and topic publication)
- [ ] T007 [P] [US1] Create integration test in `tests/integration/analyze_test.go` (using Testcontainers for Kafka)

### Implementation for User Story 1

- [ ] T008 [US1] Implement `PostAnalyze` handler in `internal/app/ingestor/service.go`
- [ ] T009 [US1] Add logic to publish message to `AnalysisTopicName` using existing Kafka producer
- [ ] T010 [US1] Ensure payload validation matches `Interaction` schema
- [ ] T010a [US1] Verify project builds successfully (`go build ./cmd/ingestor/...`)

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T011 [P] Run `make lint-go` and fix issues
- [ ] T012 Verify quickstart guide works against local environment
- [ ] T013 Verify GitHub Actions CI build success
