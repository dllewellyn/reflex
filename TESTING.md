# Unit Testing Documentation

## Overview

This repository includes comprehensive unit tests that cover functionality end-to-end. By default, these tests use in-memory implementations of external dependencies (Kafka, GCS, and Vertex AI) to validate system behavior without infrastructure. 

However, the **Ingestor** service tests now support a hybrid mode, allowing you to run the same tests against a **real Kafka cluster** by simply setting an environment variable.

## Test Structure

### Feature-Based Tests

All end-to-end unit tests are located in the `features/` directory and correspond to the BDD feature files:

- **`ingestor_unit_test.go`** - Tests for the Ingestor HTTP service. **Supports Real Kafka.**
- **`loader_unit_test.go`** - Tests for the Loader Kafka-to-GCS archival service.
- **`batch_unit_test.go`** - Tests for the Batch Analyzer LLM security scanning service.
- **`kafka_infra_test.go`** - Infrastructure helpers that enable switching between in-memory and real Kafka backends.

### In-Memory Implementations

The following in-memory implementations are available for fast, isolated testing:

1. **Kafka** (`internal/platform/kafka/memory.go`)
   - `MemoryProducer` - In-memory message producer with inspection capabilities.
   - `MemoryConsumer` - In-memory message consumer with seeding support.

2. **GCS** (`internal/platform/gcs/memory.go`)
   - `MemoryClient` - In-memory blob storage.

3. **Vertex AI** (`internal/platform/vertex/memory.go`)
   - `MemoryClient` - In-memory batch prediction client.

## Running Tests

### Install Required Tools

Before running tests or linting, install the required tools:

```bash
make tools
```

This installs `go-jsonschema`, `wire`, and `golangci-lint`.

### Generate Code

Generate schemas and wire code before testing:

```bash
make generate-go
```

### Run All Tests (In-Memory Default)

```bash
make test-go
```

Or directly:

```bash
go test ./...
```

### Run Feature Tests Only

```bash
go test ./features/... -v
```

### Run Tests with Real Kafka

To verify the **Ingestor** service against a real Kafka cluster (e.g., Confluent Cloud, Redpanda, or local Docker):

1.  Ensure your `.env` file is populated with valid credentials (see `.env.example`).
    *   `KAFKA_BOOTSTRAP_SERVERS`
    *   `KAFKA_API_KEY` (if SASL)
    *   `KAFKA_API_SECRET` (if SASL)
2.  Run the specific tests with the `TEST_USE_REAL_KAFKA` environment variable:

```bash
TEST_USE_REAL_KAFKA=true go test -v ./features/ingestor_unit_test.go ./features/kafka_infra_test.go
```

*Note: The test will use a unique consumer group and wait for the message to actually appear on the topic.*

### Run Linter

```bash
make lint-go
```

## Test Coverage

### Ingestor Service Tests

Based on `features/ingestor.feature`:

1. **TestIngestorService_ValidInteraction** ✅
   - Validates successful ingestion of valid interactions.
   - **Hybrid:** Verifies publication to either `MemoryProducer` or Real Kafka topic.

2. **TestIngestorService_InvalidJSON** ✅
   - Validates rejection of malformed JSON payloads.
   - **Hybrid:** Verifies no messages are published.

3. **TestIngestorService_MissingRequiredFields** ✅
   - Documents current behavior with incomplete payloads.

4. **TestIngestorService_KafkaUnavailable** ✅
   - Validates error handling when Kafka is unavailable.
   - *Note: This specific test uses a dedicated mock and always runs in-memory.*

### Loader Service Tests

Based on `features/loader.feature`:

1. **TestLoaderService_ArchiveMessagesToGCS** ✅
   - Validates archival of 100 messages from Kafka to GCS.
   - Verifies proper JSONL file creation.

2. **TestLoaderService_HandleZeroMessages** ✅
   - Validates graceful handling of empty topics.

3. **TestLoaderService_GCSFailure** ✅
   - Validates data consistency on GCS failures.

### Batch Analyzer Tests

Based on `features/llm-judge.feature`:

1. **TestBatchService_ProcessDailyConversations** ✅
   - Validates processing of multiple conversations from GCS.

2. **TestBatchService_ConversationsSpanningFiles** ✅
   - Validates handling of conversations split across hourly files.

3. **TestBatchService_HighRiskAlert** ✅
   - Infrastructure test for high-risk alert processing.

4. **TestBatchService_LowRiskNoAlert** ✅
   - Validates that low-risk findings don't trigger alerts.

## Test Architecture

### Hybrid Infrastructure Pattern

The `ingestor_unit_test.go` uses a `TestInfrastructure` interface to abstract the backend:

```go
type TestInfrastructure interface {
    GetProducer() kafka.Producer
    StartConsumer(t *testing.T, topic string)
    VerifyMessagePublished(t *testing.T, topic string, expectedID string)
    VerifyNoMessagePublished(t *testing.T, topic string)
    Close()
}
```

- **MemoryInfrastructure**: Instant, synchronous checks against in-memory arrays.
- **RealKafkaInfrastructure**: Asynchronous checks using a real Kafka consumer with timeouts.

### Dependency Injection

All services accept interfaces rather than concrete implementations, making them testable:

```go
// Ingestor Service
ingestor.NewService(producer kafka.Producer, cfg Config)
```

## Benefits of Hybrid Testing

1.  **Fast Feedback**: Developers run in-memory tests locally (~23ms execution).
2.  **Real Integration**: CI or ad-hoc runs can verify actual Kafka connectivity and configuration.
3.  **Code Reuse**: The same test logic validates both mocks and real integration.