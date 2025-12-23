# Research: Real-time Analysis Pipeline

**Feature**: `007-analyze-to-topic`
**Date**: 2025-12-19

## 1. Synchronous vs Asynchronous Analysis

### Decision
**Asynchronous Processing (Kafka)**

### Rationale
*   **Latency**: Security analysis (LLM probes, vector search) can take seconds. Blocking the ingest API would violate the <100ms latency requirement.
*   **Decoupling**: The ingestor should not know about the specific probes running downstream.
*   **Resilience**: Kafka acts as a buffer during traffic spikes or downstream outages.

### Alternatives Considered
*   **Synchronous HTTP**: Rejected due to high latency impact on the client.
*   **Cloud Tasks**: Viable, but Kafka is already established in the architecture (`/ingest` flow).

## 2. API Design

### Decision
**POST /analyze** with `Interaction` schema.

### Rationale
*   Consistency with `/ingest` endpoint.
*   Reusing the `Interaction` schema ensures downstream consumers receive a standardized format.

## 3. Infrastructure

### Decision
**Google Cloud Run + Kafka (Confluent/Self-hosted)**

### Rationale
*   Existing architecture uses Cloud Run for compute and Kafka for messaging.
*   Scaling capabilities of Cloud Run match the high-throughput requirement.
