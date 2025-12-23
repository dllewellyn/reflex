# Tasks: LLM Judge Batch Job

**Branch**: `003-llm-judge-batch` | **Date**: 2025-12-14 | **Plan**: [specs/003-llm-judge-batch/plan.md](specs/003-llm-judge-batch/plan.md)
**Spec**: [specs/003-llm-judge-batch/spec.md](specs/003-llm-judge-batch/spec.md)

## Phase 1: Setup
*Goal: Initialize project structure and dependencies.*

- [x] T001 [P] Create directory `prompts/` and add `security-judge.prompt.yml` template in `prompts/security-judge.prompt.yml`
- [x] T002 [P] Initialize `cmd/batch-job/` with `main.go` scaffold if not present in `cmd/batch-job/main.go`
- [x] T003 [P] Add dependencies: `cloud.google.com/go/aiplatform/apiv1` and `gopkg.in/yaml.v3` in `go.mod`
- [x] T004 [P] Create `internal/app/batch` package structure in `internal/app/batch/`
- [x] T005 [P] Create `internal/platform/vertex` package structure in `internal/platform/vertex/`

## Phase 2: Foundational
*Goal: Implement core utilities and blocking prerequisites.*

- [x] T006 [P] Implement `PromptLoader` struct and `LoadPrompt` function to parse YAML in `internal/app/batch/prompt.go`
- [x] T007 [P] Create `VertexClient` interface and wrapper struct for `JobClient` in `internal/platform/vertex/client.go`
- [x] T008 [P] Implement `VertexClient.CreateBatchJob` method wrapping the SDK call in `internal/platform/vertex/client.go`
- [x] T009 [P] Create mock for `VertexClient` to enable unit testing in `internal/platform/vertex/mock_client.go`

## Phase 3: Prompt Loading
*Goal: Enable the system to read and validate the security judge prompt.*

- [x] T010 [US1] Create unit test for `LoadPrompt` validating YAML parsing and placeholder extraction in `internal/app/batch/prompt_test.go`
- [x] T011 [US1] Implement template substitution logic (replacing `{{conversation_transcript}}`) in `internal/app/batch/prompt.go`
- [x] T012 [US1] Verify `security-judge.prompt.yml` contains required `system` and `user` roles in `prompts/security-judge.prompt.yml`

## Phase 4: Job Submission Trigger
*Goal: Trigger a Vertex AI Batch Prediction job with correct configuration.*

- [x] T013 [US2] Create unit test for `TriggerBatchJob` using mocked `VertexClient` in `internal/app/batch/trigger_test.go`
- [x] T014 [US2] Implement `TriggerBatchJob` function to construct `BatchPredictionJob` request in `internal/app/batch/trigger.go`
- [x] T015 [US2] Map `gemini-2.5-flash` model ID and GCS URIs from config to request in `internal/app/batch/trigger.go`
- [x] T016 [US2] Integrate `PromptLoader` and `TriggerBatchJob` into `cmd/batch-job/main.go`

## Phase 5: Polish
*Goal: Final cleanup and integration checks.*

- [x] T017 [P] Ensure logging uses `slog` for job submission status in `internal/app/batch/trigger.go`
- [x] T018 [P] Add comments documenting the `BatchPredictionJob` configuration fields in `internal/app/batch/trigger.go`
- [x] T019 [P] verify all files are formatted with `gofmt`
- [x] T020 Run `go mod tidy` to clean up dependencies

## Phase 6: Infrastructure & Deployment

*Goal: Deploy the solution to Google Cloud Platform.*



- [x] T021 [Infra] Create Terraform definitions for Artifact Registry, Service Accounts, and IAM roles in `terraform/llm-judge-batch.tf`

- [x] T022 [Infra] Create Terraform definitions for Cloud Run Jobs and Cloud Scheduler in `terraform/llm-judge-batch.tf`

- [x] T023 [App] Refactor `cmd/batch-job/main.go` to support Environment Variables for configuration

- [x] T024 [Docker] Create `build/Dockerfile.loader` and `build/Dockerfile.batch-job`

- [x] T025 [Docker] Build and push Docker images to Artifact Registry `flash-cache-repo` (Built via Cloud Build)

- [x] T026 [Infra] Apply Terraform configuration to deploy Cloud Run Jobs and Schedulers



*Note: `deletion_protection` was temporarily set to `false` for Cloud Run Jobs to resolve a tainted state during initial deployment.*



## Dependencies



1.  **US1 (Prompt Loading)**: Depends on T001, T006.

2.  **US2 (Job Submission)**: Depends on US1 (for prompt content), T007, T008.



## Implementation Strategy



1.  **MVP Scope**: Complete US1 and US2 to demonstrate end-to-end triggering from a CLI command.

2.  **Testing**: Unit tests for prompt parsing and job submission mocking are critical as we cannot easily trigger real Vertex jobs in CI.


