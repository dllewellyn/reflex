# Feature Specification: Real-time Analysis API

**Feature Branch**: `007-analyze-to-topic`
**Status**: Draft
**Created**: 2025-12-19

## 1. Overview

This feature introduces a new HTTP endpoint `/analyze` designed to accept interaction data for real-time security analysis. Following the architectural pattern of the existing `/ingest` endpoint, this service acts as an entry point, validating incoming requests and publishing them to a dedicated Kafka topic (`analysis-requests`) for asynchronous processing by downstream security probes.

## 2. User Scenarios

### 2.1. Real-time Security Check
**Actor**: Client Application (e.g., Chatbot Backend)
**Flow**:
1. The client application receives a message from a user.
2. The client application sends a `POST` request to `/analyze` with the interaction details (content, timestamp, IDs).
3. The system validates the payload.
4. The system queues the message for analysis.
5. The system returns an acknowledgement (`202 Accepted`) to the client, allowing the client to proceed without blocking on the full analysis.

## 3. Functional Requirements

### 3.1. API Endpoint
*   **FR-001**: The application MUST expose a `POST /analyze` endpoint.
*   **FR-002**: The endpoint MUST accept a JSON payload compliant with the `Interaction` event schema (see `specifications/schemas/interaction-event.schema.json`).
*   **FR-003**: The endpoint MUST return `400 Bad Request` if the payload fails schema validation.

### 3.2. Data Processing
*   **FR-004**: Upon successful validation, the application MUST publish the event to the `analysis-requests` Kafka topic.
*   **FR-005**: The published message MUST contain the exact payload received in the request.

### 3.3. Response Handling
*   **FR-006**: The endpoint MUST return `202 Accepted` immediately after successfully publishing the message to Kafka.
*   **FR-007**: The endpoint MUST return `500 Internal Server Error` if the message cannot be published to Kafka (e.g., broker unavailable).

## 4. Success Criteria

*   **SC-001**: **Throughput**: The endpoint successfully handles and acknowledges valid requests within 100ms (P99) under normal load.
*   **SC-002**: **Reliability**: 99.9% of valid requests received are successfully published to the Kafka topic.
*   **SC-003**: **Validation**: 100% of malformed requests (invalid JSON, missing required fields) are rejected with a 4xx error code.

## 5. Assumptions

*   The `Interaction` schema (`specifications/schemas/interaction-event.schema.json`) defines the correct data structure for analysis requests.
*   The `analysis-requests` Kafka topic will be provisioned as part of the infrastructure setup or exists.
*   Authentication/Authorization is handled at the gateway or service level (out of scope for this specific functional spec, assuming consistent with `/ingest`).