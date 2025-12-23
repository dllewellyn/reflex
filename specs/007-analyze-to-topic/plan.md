# Implementation Plan: Analyze to Topic

**Branch**: `007-analyze-to-topic` | **Date**: 2025-12-19 | **Spec**: [specs/007-analyze-to-topic/spec.md](spec.md)
**Input**: Feature specification from `specs/007-analyze-to-topic/spec.md`

## Summary

This feature implements the `/analyze` endpoint in the Ingestor service to support real-time security analysis of user/model interactions. It involves updating the OpenAPI specification to include the new endpoint, regenerating the server code, implementing the handler to validate and publish payloads to a new `analysis-requests` Kafka topic, and provisioning the necessary infrastructure.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: `github.com/oapi-codegen/runtime` (OpenAPI), `github.com/segmentio/kafka-go` (Kafka)
**Storage**: Kafka (Topic: `analysis-requests`)
**Testing**: Go `testing` package (Unit), `testcontainers-go` (Integration)
**Target Platform**: Google Cloud Run
**Project Type**: Backend Service
**Performance Goals**: <100ms P99 latency (excluding Kafka RTT)
**Constraints**: Must match existing `/ingest` patterns; schema validation is critical.
**Scale/Scope**: High throughput potential; relies on Kafka for buffering.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **BDD Tests:** All features must have BDD test cases defined in this plan.
- [x] **Observability:** Logging and OpenTelemetry integration must be planned.
- [x] **E2E Testing:** Plan includes testing against real infrastructure (Kafka/Firebase) in CI.
- [x] **Infrastructure:** Infrastructure requirements identified?
- [x] **Serverless:** All new infrastructure is serverless/scale-to-zero OR explicitly justified?
- [x] **Secrets:** Secrets management strategy defined?
- [x] **Schemas:** JSON Schemas/OpenAPI defined for external interfaces?
- [x] **Codegen:** Plan includes automated code generation from schemas?

## BDD Test Cases

*As a client application, I can send interaction data for analysis so that I can receive real-time security feedback.*

| Feature | Scenario | Given | When | Then |
| --- | --- | --- | --- | --- |
| Real-time Analysis | Successful Submission | A valid interaction JSON payload | I POST to `/analyze` | I receive `202 Accepted` AND the message is published to `analysis-requests` topic |
| Real-time Analysis | Invalid Schema | A JSON payload missing required fields | I POST to `/analyze` | I receive `400 Bad Request` AND no message is published |
| Real-time Analysis | Kafka Failure | A valid interaction JSON payload AND Kafka is unreachable | I POST to `/analyze` | I receive `500 Internal Server Error` |

## Project Structure

### Documentation (this feature)

```text
specs/007-analyze-to-topic/
├── plan.md              # This file
├── research.md          # Architecture decisions
├── data-model.md        # Data schemas
├── contracts/           # API definitions
├── tasks.md             # Task breakdown
└── checklists/          # Verification checklists
```

### Source Code (repository root)

```text
specifications/
└── ingestor-api.yaml    # OpenAPI definition update

cmd/ingestor/
├── main.go              # Main entry point (wiring)
└── wire.go              # Dependency injection

internal/app/ingestor/
├── api.gen.go           # Generated code (updated)
└── service.go           # Handler implementation

terraform/
├── ingestor.tf          # Environment variable updates
└── variables.tf         # New topic variable
```

## Implementation Strategy

1.  **Specification & Codegen**: Update `specifications/ingestor-api.yaml` to define `/analyze` with the `Interaction` schema. Regenerate Go code.
2.  **Infrastructure**: Update Terraform to define the `analysis-requests` topic and pass it as an environment variable (`ANALYSIS_TOPIC`) to the Cloud Run service.
3.  **Application Logic**: Implement the `PostAnalyze` handler in `internal/app/ingestor`.
    *   Reuse the existing Kafka producer mechanism if possible, or refactor to support multiple topics.
    *   Validate payload (handled mostly by OAPI middleware/types).
    *   Publish to the new topic.
4.  **Testing**:
    *   Unit tests for the handler.
    *   Integration tests using Testcontainers (Kafka) to verify end-to-end flow.
5.  **Deployment**: Apply Terraform changes and deploy the updated service.

## Complexity Tracking

No violations. Reusing existing patterns for consistency.