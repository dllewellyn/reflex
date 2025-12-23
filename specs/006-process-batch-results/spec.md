# Feature Specification: Automate Batch Results Processing

**Feature Branch**: `006-process-batch-results`
**Created**: 2025-12-18
**Status**: Draft
**Input**: User description: "Update our existing logic so when results are processed into our batch results GCS bucket, we are processing the jsonl results file automatically (using a GCS storage trigger) and serverless to then push each jsonl line, one at a time into kafka. Existing logic for `cmd/extract-injections` will then be updated to process these (one at a time) instead of the current logic to pull"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic Batch Ingestion (Priority: P1)

As a system operator, I want new batch result files to be automatically processed as soon as they are uploaded to storage, so that I don't have to manually trigger ingestion.

**Why this priority**: Automation is the primary goal of this feature to reduce manual toil and latency.

**Independent Test**: Upload a valid JSONL file to the designated GCS bucket and verify that the processing logic is triggered automatically.

**Acceptance Scenarios**:

1. **Given** a configured GCS bucket for batch results, **When** a new JSONL file is uploaded, **Then** the serverless function is triggered to process the file.
2. **Given** a non-JSONL file is uploaded (if applicable), **When** the trigger fires, **Then** the system handles it gracefully (ignores or logs error).

---

### User Story 2 - Streaming Results to Kafka (Priority: P1)

As a downstream consumer, I want batch results to be split into individual messages on a Kafka topic, so that I can process each result independently and in parallel.

**Why this priority**: Enables the architectural shift from batch pull to event-driven push/stream processing.

**Independent Test**: Trigger the processing function with a sample JSONL file and verify that the expected number of messages appear in the target Kafka topic, with correct content.

**Acceptance Scenarios**:

1. **Given** a JSONL file with 100 records, **When** the file is processed, **Then** 100 distinct messages are published to the Kafka topic.
2. **Given** a JSONL file with malformed lines, **When** the file is processed, **Then** valid lines are published and invalid lines are logged/handled.

---

### User Story 3 - Stream Consumer Update (Priority: P1)

As a system developer, I want the `extract-injections` component to consume results from Kafka instead of pulling from GCS, so that it integrates with the new event-driven pipeline.

**Why this priority**: Required to close the loop and actually use the data being streamed.

**Independent Test**: Publish messages to the Kafka topic and verify `extract-injections` processes them correctly.

**Acceptance Scenarios**:

1. **Given** messages in the Kafka topic, **When** `extract-injections` is running, **Then** it consumes and processes the messages one by one.
2. **Given** the new pipeline is active, **When** a batch is finished, **Then** `extract-injections` has processed all records from that batch.

### Edge Cases

- **Huge Files**: What happens when the JSONL file is larger than the serverless function's memory or execution time limit? (Streaming reading required).
- **Kafka Down**: What happens if Kafka is unavailable when the function tries to publish? (Retries/Failure handling).
- **Duplicate Events**: What happens if the GCS trigger fires multiple times for the same file? (Idempotency considerations).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST automatically detect the creation of new files in the designated Batch Results GCS bucket.
- **FR-002**: System MUST execute a serverless function (e.g., Cloud Function) in response to the file creation event.
- **FR-003**: The serverless function MUST read the content of the detected JSONL file.
- **FR-004**: The system MUST support streaming reading of the file to handle files larger than available memory.
- **FR-005**: The system MUST parse each line of the JSONL file as a distinct record.
- **FR-006**: The system MUST publish each parsed record as a separate message to a configured Kafka topic.
- **FR-007**: The `extract-injections` component MUST be updated to consume messages from the Kafka topic.
- **FR-008**: The `extract-injections` component MUST NO LONGER pull files directly from GCS for processing (legacy logic removal/update).
- **FR-009**: The system MUST log any errors encountered during parsing or publishing.

### Key Entities

- **Batch Result File**: The source object in GCS containing multiple result records in JSONL format.
- **Result Record**: A single unit of data extracted from a line in the Batch Result File.
- **Ingestion Event**: The message on Kafka representing a single Result Record ready for processing.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New batch result files are picked up for processing within 1 minute of upload completion.
- **SC-002**: 100% of valid records in a batch file are successfully published to Kafka.
- **SC-003**: `extract-injections` successfully processes messages from Kafka with equivalent or better throughput than the previous pull-based mechanism.
- **SC-004**: The system handles files up to 10 GB without running out of memory.