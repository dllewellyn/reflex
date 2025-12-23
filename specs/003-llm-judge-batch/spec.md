# System Specification: LLM As A Judge - Scheduled Batch Analysis

**Version**: 5.0
**Status**: Draft
**Created**: December 12, 2025
**Pivot**: Flink -> Cron-based Batch Loading

## 1. Overview

The system uses a **Scheduled Batch Architecture** to detect prompt injection.
1.  **Ingestion**: API receives interactions and pushes to Kafka.
2.  **Archival**: An **Hourly Loader Job** consumes Kafka and dumps raw data to GCS.
3.  **Analysis**: A **Daily Batch Job** aggregates the raw data into conversations and submits them to **Vertex AI (Gemini Flash 2.5)** for security evaluation.

## 2. Architecture

### High-Level Data Flow

1.  **Ingestor Service**: REST API (`POST /ingest`) -> Kafka (`raw-interactions`).
2.  **Loader Job (Hourly)**:
    *   Triggered by Cloud Scheduler.
    *   Consumes all new messages from `raw-interactions`.
    *   Writes a single JSONL blob to `gs://<bucket>/raw/<conversation_id>/YYYY/MM/DD/HH/`.
3.  **Batch Analyzer (Daily)**:
    *   Triggered by Cloud Scheduler.
    *   Identifies all sessions that had activity on the target date.
    *   Retrieves the **full conversation history** for those sessions (spanning all dates).
    *   **Groups messages by `conversation_id`** (reconstructing sessions).
    *   Formats and submits to **Vertex AI Batch Prediction**.
4.  **Result Streaming**:
    *   Analyzer parses Vertex results.
    *   High-risk findings are pushed to Kafka (`security-alerts`).

### Components

*   **Ingestor (Go)**: Continuous Service. API Gateway.
*   **Kafka**: Durable Buffer.
*   **Loader (Go)**: Scheduled Task. Kafka -> GCS Sink.
*   **Batch Analyzer (Go)**: Scheduled Task. GCS -> Vertex AI -> Kafka.

## 3. Data Model

### Interaction Event (Kafka & GCS Raw)

```json
{
  "interaction_id": "uuid",
  "conversation_id": "uuid",
  "timestamp": "iso8601",
  "role": "user|model",
  "content": "string"
}
```

### Batch Input (Vertex AI)

Grouped by the Daily Analyzer.

```json
{
  "request": {
    "contents": [
      {
        "role": "user",
        "parts": [{ "text": "<JUDGE_PROMPT> ... <CONVERSATION_TRANSCRIPT>" }]
      }
    ]
  }
}
```

## 4. Requirements

### Functional
*   **FR-001**: Loader MUST reliably consume from Kafka and commit offsets only after GCS write.
*   **FR-002**: Batch Analyzer MUST handle conversations that span across hourly files.
*   **FR-003**: System MUST provide feedback on high-severity threats via Kafka.

### Non-Functional
*   **NFR-001**: Loader should handle "zero messages" gracefully.
*   **NFR-002**: Batch Analyzer must fit daily volume in memory (or use streaming aggregation if >1GB).