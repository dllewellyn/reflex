# Data Model

## Entities

### Batch Result Record
*Represents a single line from the Vertex AI Batch Prediction output file.*

- **Structure**: JSON object.
- **Fields** (Standard Vertex AI Output):
  - `instance`: The input instance (prompt/context).
  - `prediction`: The model's output.
  - `status`: Success/Failure string.

### Batch Result Event (Kafka Message)
*The event published to the `batch-job-results` topic.*

- **Schema Name**: `batch-result-event.schema.json`
- **Fields**:
  - `eventId`: UUID (String).
  - `timestamp`: ISO8601 (String).
  - `source`:
    - `bucket`: String.
    - `file`: String.
  - `record`: Object (The `Batch Result Record` above).

## Relationships

- One **Batch Result File** (GCS) contains Many **Batch Result Records**.
- One **Batch Result Record** corresponds to One **Batch Result Event** (Kafka).
