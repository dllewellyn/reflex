# Research & Architecture Decisions

**Feature**: Automate Batch Results Processing
**Branch**: `006-process-batch-results`

## Decisions

### 1. Kafka Topic Naming & Configuration
- **Decision**: Use topic name `batch-job-results`.
- **Rationale**: Clear, descriptive, and follows the pattern of mapping data sources to topics.
- **Configuration**: Partition count should match the expected concurrency of the consumers (e.g., `extract-injections`). Default to 4 partitions for now.

### 2. Kafka Message Format
- **Decision**: Use a structured JSON schema wrapper.
- **Schema**: `BatchResultEvent`
    - `metadata`: Contains `source_bucket`, `source_file`, `processed_at`.
    - `payload`: The actual JSON object from the batch result line.
- **Rationale**: Raw JSON lines from Vertex AI might lack context (which file did they come from?). Wrapping them adds traceability, which is crucial for debugging and potential reprocessing.

### 3. Deployment Mechanism
- **Decision**: Use Terraform (`google_cloudfunctions2_function`).
- **Rationale**: The project already manages infrastructure (Cloud Run, Scheduler, GCS) via Terraform (`terraform/` directory). Mixing manual deployment with Terraform is an anti-pattern (Principle XII).

### 4. Cloud Function Runtime
- **Decision**: Go 1.23+.
- **Rationale**: Consistent with the rest of the backend services (Principle IV).

## Resolved Unknowns

| Unknown | Resolution | Source |
| :--- | :--- | :--- |
| Kafka Topic Name | `batch-job-results` | Decision 1 |
| Message Format | Wrapped `BatchResultEvent` | Decision 2 |
| Deployment Tool | Terraform | Decision 3 |

## Alternatives Considered

- **Direct GCS to Kafka Connector**:
    - *Pros*: No code to write.
    - *Cons*: Less control over validation and transformation. Might require running a Kafka Connect cluster (violates Serverless principle if not managed).
- **Pub/Sub instead of Kafka**:
    - *Pros*: Native GCS integration.
    - *Cons*: Project explicitly uses Kafka (`001-gcp-kafka-ingest`). Mixing messaging backbones adds complexity.
