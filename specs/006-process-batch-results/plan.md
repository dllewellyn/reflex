# Implementation Plan - Automate Batch Results Processing

**Feature Branch**: `006-process-batch-results`
**Feature Spec**: `specs/006-process-batch-results/spec.md`

## Technical Context

### Existing Logic
- `cmd/batch-job` (Go):
    - Submits Vertex AI Batch Prediction jobs.
    - Input: `gs://<staging_bucket>/staging/<YYYY>/<MM>/<DD>/*.jsonl`
    - Output: `gs://<processed_bucket>/results/<YYYY>/<MM>/<DD>/` (Batch Prediction writes results here).
- `cmd/extract-injections` (Go):
    - Currently configured via `extract.Config`.
    - Likely pulls/reads files directly from GCS (logic to be verified/replaced).
- Infrastructure (`terraform/llm-judge-batch.tf`):
    - `google_storage_bucket.processed_prompts`: The bucket where batch results land.
    - `google_cloud_run_v2_job.llm_judge_trigger`: The daily batch job.

### Proposed Changes
1.  **Infrastructure**:
    - Add a Google Cloud Function (2nd Gen) triggered by `google.storage.object.finalize` on `google_storage_bucket.processed_prompts`.
    - Grant necessary IAM roles (Storage Object Viewer, Cloud Run Invoker, etc.).
    - Configure the function with Kafka connection details.

2.  **New Component (Cloud Function)**:
    - Language: Go 1.23+ (consistent with project).
    - Logic:
        - Triggered by GCS event.
        - Validate file name pattern (ensure it's a result file).
        - Stream read the file from GCS.
        - Parse JSONL.
        - Publish each record to Kafka.

3.  **Update `cmd/extract-injections`**:
    - Remove GCS polling/reading logic.
    - Add Kafka consumer logic.
    - Update configuration to accept Kafka connection details instead of/in addition to GCS buckets.

### Unknowns & Clarifications
- **[RESOLVED] Kafka Topic Configuration**: Topic name `batch-job-results`.
- **[RESOLVED] Kafka Message Format**: Wrapped `BatchResultEvent`.
- **[RESOLVED] Cloud Function deployment**: Terraform.

## Constitution Check

- **[x]** BDD Test Cases in Plan Phase (Principle I)
- **[x]** Serverless & Scale-to-Zero (Principle XII) - Cloud Function fits this perfectly.
- **[x]** Infrastructure Planning (Principle XIII) - New Cloud Function and IAM bindings required.
- **[x]** Schema-Driven Development (Principle XIV) - Schema `batch-result-event.schema.json` created.

## Phase 0: Research & Decisions

- See `specs/006-process-batch-results/research.md` for detailed decisions on Kafka topics, message formats, and deployment strategy.

## Phase 1: Design & Contracts

- **Data Model**: Defined in `specs/006-process-batch-results/data-model.md`.
- **Contracts**: JSON Schema created at `specifications/schemas/batch-result-event.schema.json`.
- **Agent Context**: Updated `GEMINI.md`.

## Phase 2: Implementation Steps

1.  **Generate Schema Code**:
    - Run `make generate` (or equivalent script) to generate Go structs from `specifications/schemas/batch-result-event.schema.json`.
    - **Verification**: Check for new generated files in `internal/platform/schema` (or configured output).

2.  **Implement Cloud Function (Trigger)**:
    - Create `cmd/batch-result-trigger/main.go`.
    - Implement GCS event handler.
    - Implement streaming JSONL reader.
    - Implement Kafka producer using the generated schema.
    - **Test**: Unit tests for parser and handler.

3.  **Update Infrastructure (Terraform)**:
    - Edit `terraform/llm-judge-batch.tf` (or new file) to define `google_cloudfunctions2_function`.
    - Add IAM roles for the function SA.
    - **Verification**: `terraform plan` shows correct additions.

4.  **Update `extract-injections` Consumer**:
    - Modify `internal/app/extract/config.go` to add Kafka config.
    - Implement/Update Kafka consumer in `internal/app/extract/`.
    - Wire up the new consumer in `cmd/extract-injections/wire.go`.
    - Remove legacy GCS pull logic.
    - **Test**: Integration test with a local Kafka and producer.

5.  **End-to-End Test**:
    - Deploy infrastructure.
    - Upload sample file to GCS.
    - Verify Cloud Function logs.
    - Verify `extract-injections` logs processing the event.
