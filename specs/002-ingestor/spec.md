# Feature Specification: REST API Ingestor

**Feature Branch**: `002-ingestor`
**Status**: Active
**Created**: December 12, 2025

## 1. Overview

The **Ingestor Service** is a high-throughput HTTP API that acts as the entry point for the security analysis pipeline. It accepts JSON payloads representing user/AI interactions and publishes them to a Kafka topic for asynchronous processing.

## 2. Requirements

### Functional

* **FR-001**: The Ingestor MUST expose a `POST /ingest` endpoint.
* **FR-002**: The Ingestor MUST validate incoming JSON against the `Interaction` schema.
* **FR-003**: Valid messages MUST be published to the `raw-interactions` Kafka topic.
* **FR-004**: The Ingestor MUST return `202 Accepted` upon successful handoff to Kafka.
* **FR-005**: If Kafka is unreachable, the Ingestor MUST return `500 Internal Server Error`.
* **FR-006**: The ingestor must have a fully defined openapi.yaml spec and the server code is generated from it.

### Non-Functional

* **NFR-001**: The service must handle concurrent requests efficiently.
* **NFR-002**: Graceful shutdown must be implemented to finish in-flight requests.

## 3. Success Criteria

* **SC-001**: 99.9% of valid requests are successfully published to Kafka.
* **SC-002**: P99 Latency < 100ms (excluding Kafka RTT).
