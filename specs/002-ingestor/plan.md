# Implementation Plan: REST API Ingestor

**Branch**: `002-ingestor` | **Spec**: [spec.md](./spec.md)

## Summary

Implement a Go HTTP Service (`cmd/ingestor`) that wraps a Kafka Producer.

## Architecture

*   **Runtime**: Go 1.24 (Cloud Run Service)
*   **Input**: HTTP `POST /ingest`
*   **Output**: Kafka Topic `raw-interactions`

## Steps

1.  **Project Setup**: Create `cmd/ingestor` and `internal/app/ingestor`.
2.  **Schema**: Define `Interaction` struct in `internal/platform/schema`.
3.  **Producer**: Implement `KafkaProducer` wrapper in `internal/platform/kafka`.
4.  **Handler**: Implement HTTP handler to Unmarshal -> Validate -> Produce.
5.  **Wiring**: Use `wire` to inject Producer into Handler.
