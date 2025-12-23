# Feature Specification: Hourly Loader Job

**Feature Branch**: `001-loader`
**Status**: Active
**Created**: December 12, 2025

## 1. Overview

The **Loader Job** is a scheduled background process that runs hourly. Its primary responsibility is to consume all available messages from a Kafka topic (`raw-interactions`) and archive them into Google Cloud Storage (GCS) in a raw JSONL format. This creates a durable data lake for downstream batch analysis.

## 2. Requirements

### Functional

* **FR-001**: The Loader MUST start, consume messages until the topic is empty (or a timeout occurs), and then exit ("Run-Once" semantics).
* **FR-002**: The Loader MUST buffer messages in memory and write them as a single object to GCS to minimize file fragmentation.
* **FR-003**: The GCS object key MUST follow the pattern `raw/<chat_session_id>/YYYY/MM/DD/HH/chunk-<uuid>.jsonl`.
* **FR-004**: The Loader MUST commit Kafka offsets *only after* a successful GCS write to ensure at-least-once delivery.
* **FR-005**: The loader MUST retrieve the Chat Session UUID from the kafka message and use it as the GCS object key prefix.

### Non-Functional

* **NFR-001**: The job must handle execution timeouts gracefully.
* **NFR-002**: Configuration (Topic, Bucket) must be loaded from environment variables.

## 3. Success Criteria

* **SC-001**: All messages published to Kafka are eventually visible in GCS.
* **SC-002**: Zero data loss during job restarts.
