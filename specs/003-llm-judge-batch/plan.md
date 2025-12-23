# Implementation Plan: LLM Judge Batch Job

**Branch**: `003-llm-judge-batch` | **Date**: 2025-12-14 | **Spec**: [specs/003-llm-judge-batch/spec.md](specs/003-llm-judge-batch/spec.md)
**Input**: Feature specification from `/specs/003-llm-judge-batch/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement the triggering logic for a Vertex AI Batch Prediction job using `gemini-2.5-flash`.
Key components:
1.  **Prompt Management**: Load the judge prompt from `prompts/security-judge.prompt.yml` (GitHub Models format).
2.  **Job Submission**: Use `aiplatform` Go client to submit the batch job.
3.  **Input/Output**: Configure GCS source (JSONL) and destination.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**:
- `cloud.google.com/go/aiplatform/apiv1` (Job Submission)
- `gopkg.in/yaml.v3` (Parsing .prompt.yml)
**Storage**: Google Cloud Storage (Source & Sink)
**Testing**: Go `testing`, Mocking `JobClient`
**Target Platform**: Kubernetes / Cloud Run Jobs (Linux)
**Project Type**: Batch Job (CLI)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **BDD Tests:** All features must have BDD test cases defined in this plan.
- [x] **Observability:** Logging (slog) and OpenTelemetry integration must be planned.
- [x] **E2E Testing:** Plan includes testing against real infrastructure (Kafka/GCS emulator) in CI.

## BDD Test Cases

*As a system, I can trigger a batch job so that large-scale analysis is performed efficiently.*

| Feature | Scenario | Given | When | Then |
| --- | --- | --- | --- | --- |
| Prompt Loading | Read Prompt File | A valid `prompts/security-judge.prompt.yml` file | The Prompt Loader parses the file | It extracts the `system` instructions and `user` template with `{{conversation_transcript}}` placeholder. |
| Job Submission | Trigger Vertex Batch | A valid GCS input URI `gs://bucket/input.jsonl` | The Trigger function is called | A `BatchPredictionJob` is created in Vertex AI with state `JOB_STATE_PENDING`. |

## Project Structure

### Documentation (this feature)

```text
specs/003-llm-judge-batch/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
prompts/
└── security-judge.prompt.yml  # NEW: GitHub Models format

cmd/
└── batch-job/
    ├── main.go
    └── wire.go

internal/
├── app/
│   └── batch/
│       ├── prompt.go          # NEW: Logic to parse .prompt.yml
│       ├── trigger.go         # Vertex AI Job Submission Logic
│       └── service.go         # Orchestration
└── platform/
    └── vertex/
        └── client.go          # Wrapper around aiplatform.JobClient
```

**Structure Decision**: Option 1 (Single project) - Extending existing `cmd/batch-job` scaffold.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | | |
