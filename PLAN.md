# Implementation Plan: Scheduled Batch Evaluation

**Status**: Active
**Architecture**: Cron Loader -> GCS -> Vertex Batch

## 1. Objective

Implement a cost-effective pipeline where an hourly job archives Kafka data to GCS, and a daily job performs "LLM as a Judge" security analysis.

## 2. Technical Stack

*   **Ingestion**: Go (REST API)
*   **Storage**: Google Cloud Storage (JSONL)
*   **Evaluation**: Vertex AI Batch Prediction
*   **Orchestration**: Cloud Run Jobs (triggered by Scheduler)

## 3. Directory Structure

```text
cmd/
├── ingestor/
│   └── main.go          # HTTP -> Kafka
├── loader/
│   └── main.go          # Kafka -> GCS (Run-Once)
└── batch-job/
    └── main.go          # GCS -> Vertex -> Kafka

internal/
├── app/
│   ├── ingestor/
│   ├── loader/
│   │   └── service.go   # Consumer Logic
│   └── batch/
│       └── service.go   # Aggregation Logic
└── platform/
    ├── gcs/             # Reader/Writer
    └── kafka/           # Consumer/Producer
```

## 4. Implementation Steps

### Phase 1: Ingestion
*   **Status**: Done (`cmd/ingestor`).

### Phase 2: Hourly Loader (`cmd/loader`)
*   **Goal**: Create a job that consumes all available messages and exits.
*   **Logic**:
    1.  Start Kafka Consumer.
    2.  Read messages with a short timeout (e.g., 10s of silence = done).
    3.  Buffer in memory.
    4.  Write `chunk-<uuid>.jsonl` to `gs://bucket/raw/YYYY/MM/DD/HH/`.
    5.  Commit Offsets.
    6.  Exit.

### Phase 3: Daily Analyzer (`cmd/batch-job`)
*   **Goal**: Evaluate the day's traffic.
*   **Logic**:
    1.  List all files in `gs://bucket/raw/YYYY/MM/DD/`.
    2.  Load and **Group by `conversation_id`**.
    3.  Format for Vertex.
    4.  Submit Job.
    5.  Process Results -> Alert to Kafka.

## 5. Development Phases

1.  **Restore Loader**: Re-implement `cmd/loader` optimized for "Batch Consumption".
2.  **Batch Job**: Implement the Aggregator and Vertex Client.
