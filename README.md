# Reflex

A cloud-native security monitoring platform that detects prompt injection attacks and other security threats in AI conversation data using LLM-as-a-Judge analysis powered by Google Vertex AI.

## Overview

Reflex implements a scheduled batch architecture for cost-effective security analysis:

1. **Real-time Ingestion**: REST API receives conversation interactions and publishes to Kafka
2. **Hourly Archival**: Scheduled job consumes from Kafka and archives raw data to Google Cloud Storage
3. **Daily Analysis**: Batch job processes conversations through Vertex AI (Gemini 2.5 Flash) for security evaluation
4. **Dataset Management**: Tool for loading conversation datasets into Pinecone for similarity search and analysis

## Architecture

### Components

- **Ingestor Service** (`cmd/ingestor`): HTTP API that receives interaction events and publishes to Kafka
- **Loader Job** (`cmd/loader`): Hourly scheduled job that consumes from Kafka and archives to GCS as JSONL
- **Batch Analyzer** (`cmd/batch-job`): Daily job that processes GCS data through Vertex AI batch prediction
- **Dataset Loader** (`cmd/dataset-loader`): Utility for ingesting datasets into Pinecone vector database
- **Extract Injections** (`cmd/extract-injections`): Processes batch analysis results to extract and upsert specific prompt injection strings into Pinecone.

### Data Flow

```lisp
User -> REST API (Ingestor) -> Kafka (raw-interactions)
                                  ↓
                            Loader (Hourly)
                                  ↓
                            GCS (JSONL Archives)
                                  ↓
                         Batch Analyzer (Daily)
                                  ↓
                    Vertex AI (Gemini 2.5 Flash)
                                  ↓
                    Kafka (security-alerts)
```

## Configuration

The system is configured via environment variables. See `.env.example` for all available options.

### Core Configuration

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `GOOGLE_CLOUD_PROJECT` | GCP Project ID | Yes | - |
| `KAFKA_BOOTSTRAP_SERVERS` | Kafka broker addresses | Yes* | localhost:9092 |
| `KAFKA_API_KEY` | Kafka SASL username (Confluent Cloud) | No | - |
| `KAFKA_API_SECRET` | Kafka SASL password (Confluent Cloud) | No | - |
| `KAFKA_TOPIC` | Kafka topic for raw interactions | Yes | prompts-topic |

*Can be loaded from `client.properties` file instead

### Service-Specific Configuration

#### Ingestor

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | 8080 |
| `KAFKA_CONFIG_FILE` | Path to Kafka config | client.properties |

#### Loader

| Variable | Description | Default |
|----------|-------------|---------|
| `GCS_BUCKET` | GCS bucket for archives | Required |
| `RAW_INTERACTIONS_TOPIC` | Kafka topic to consume | raw-interactions |
| `KAFKA_CONSUMER_GROUP_ID` | Consumer group ID | prompt-injection-worker-group |

#### Batch Analyzer

| Variable | Description | Default |
|----------|-------------|---------|
| `GCP_PROJECT` | GCP Project ID | Required |
| `GCP_LOCATION` | GCP region | us-central1 |
| `GCS_RAW_PROMPT_BUCKET` | Source bucket for raw data | - |
| `GCS_BATCH_STAGING_BUCKET` | Staging bucket | - |
| `GCS_PROCESSED_PROMPT_BUCKET` | Output bucket | - |
| `MODEL_ID` | Vertex AI model | publishers/google/models/gemini-2.5-flash |
| `PROMPT_PATH` | Security judge prompt file | prompts/security-judge.prompt.yml |

#### Dataset Loader

| Variable | Description | Required |
|----------|-------------|----------|
| `PINECONE_API_KEY` | Pinecone API key | Yes |
| `HF_TOKEN` | HuggingFace API token | Yes |

## Running Services

### Ingestor (Continuous Service)

Run locally using `go run`:

```bash
go run cmd/ingestor/main.go
```

Or run the built binary:

```bash
./bin/ingestor
```

The API will be available at `http://localhost:8080`. See the OpenAPI specification in `specifications/ingestor-api.yaml` for endpoint details.

### Loader (Scheduled Job)

Run once (typically scheduled hourly):

```bash
go run cmd/loader/main.go
```

### Batch Analyzer (Scheduled Job)

Run with automatic date detection (processes yesterday's data):

```bash
go run cmd/batch-job/main.go
```

Or specify custom input/output:

```bash
go run cmd/batch-job/main.go \
  -project your-project \
  -input gs://bucket/staging/2025/12/16/*.jsonl \
  -output gs://bucket/results/2025/12/16/
```

### Dataset Loader

```bash
go run cmd/dataset-loader/main.go
```

### Extract Injections

### Extract Injections

Run with input from processed batch results (e.g., yesterday's data).
Requires `GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION`, `PINECONE_API_KEY`, `PINECONE_INDEX_HOST` to be set.
Use the `.env` file for configuration.

```bash
set -a && source .env && set +a && \
INPUT_URI="gs://${GCS_PROCESSED_PROMPT_BUCKET}/results/$(date -v-1d +%Y/%m/%d)/*.jsonl" \
./bin/extract-injections
```

Run in dry-run mode:

```bash
set -a && source .env && set +a && \
INPUT_URI="gs://${GCS_PROCESSED_PROMPT_BUCKET}/results/$(date -v-1d +%Y/%m/%d)/*.jsonl" \
DRY_RUN="true" \
./bin/extract-injections
```

## Testing

The project includes comprehensive unit tests and integration tests. Tests use in-memory implementations by default for fast feedback.

### Run All Tests

### Run All Tests

Run all unit tests:

```bash
make test-go
```

Or directly:

```bash
go test ./...
```

Run integration tests:

```bash
go test -tags=integration ./...
```

### Run Feature Tests

### Run Feature Tests

Run specific feature tests:

```bash
go test ./features/... -v
```

Run with real Kafka (requires configuration):

```bash
TEST_USE_REAL_KAFKA=true go test -v ./features/ingestor_unit_test.go ./features/kafka_infra_test.go
```

### Test Coverage

- **Ingestor**: API validation, Kafka publishing, error handling
- **Loader**: Kafka-to-GCS archival, zero-message handling, failure recovery
- **Batch Analyzer**: Conversation aggregation, Vertex AI integration, alert generation

See `TESTING.md` for detailed testing documentation.

## Development

### Prerequisites

- Go 1.24 or later
- [golangci-lint](https://golangci-lint.run/) for linting
- [wire](https://github.com/google/wire) for dependency injection
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) for API code generation
- [go-jsonschema](https://github.com/atombender/go-jsonschema) for schema generation
- Access to Google Cloud Platform
- Kafka cluster (local, Confluent Cloud, or other)
- (Optional) Pinecone account for dataset loading

### Setup

1. **Clone the repository** (Skip if you are already in the directory)

```bash
git clone https://github.com/dllewellyn/reflex.git
cd reflex
```

2. **Install required tools**

This installs all necessary development tools including wire, golangci-lint, and code generators.

```bash
make tools
```

3. **Set up environment variables**

Copy the example configuration to `.env`.

```bash
cp -n .env.example .env || echo ".env already exists, skipping copy"
```

> [!NOTE]
> You must edit `.env` with your actual configuration values (GCP Project ID, Kafka credentials, etc.) before running services.

4. **Generate code**

Generate JSON schemas, Wire dependency injection, and OpenAPI server code.

```bash
make generate-go
```

The project uses several code generation tools. Always run before building:

```bash
make generate-go
```

Individual generation commands:

- `make generate-schemas` - Generate Go types from JSON schemas
- `make generate-wire` - Generate dependency injection code
- `make generate-api` - Generate OpenAPI server code

### Formatting

Format Go code:

```bash
make fmt-go
```

Check if code is formatted correctly:

```bash
make fmt-check
```

### Linting

Run the linter:

```bash
make lint-go
```

### Building

Build all binaries:

```bash
make build-go
```

This creates binaries in the `bin/` directory:

- `bin/ingestor` - REST API service
- `bin/loader` - Hourly archival job
- `bin/batch-job` - Daily analysis job
- `bin/dataset-loader` - Dataset ingestion utility
- `bin/extract-injections` - Utility for extracting and upserting prompt injection strings

Build individual services:

- `make build-ingestor`
- `make build-loader`
- `make build-batch`
- `make build-dataset-loader`
- `make build-extract`

### Available Make Targets

| Target | Description |
|--------|-------------|
| `make tools` | Install required development tools |
| `make generate-go` | Generate all code (schemas + wire + API) |
| `make fmt-go` | Format all Go code |
| `make fmt-check` | Check if Go code is formatted |
| `make lint-go` | Lint Go code |
| `make test-go` | Run all tests (unit + integration) |
| `make build-go` | Build all binaries |
| `make build-ingestor` | Build ingestor service only |
| `make build-loader` | Build loader job only |
| `make build-batch` | Build batch analyzer only |
| `make build-dataset-loader` | Build dataset loader only |
| `make docker-build` | Build Docker images |
| `make infrastructure` | Deploy infrastructure via Terraform |
| `make validate-tf` | Validate Terraform configuration |
| `make clean` | Clean build artifacts |
| `make tidy` | Run go mod tidy |

## Infrastructure Setup

This project uses Terraform to manage Google Cloud resources. Infrastructure code is located in the `terraform/` directory.

### Prerequisites

Ensure you have the following tools installed:

- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) (`gcloud`)
- [Terraform](https://developer.hashicorp.com/terraform/install)
- [Make](https://www.gnu.org/software/make/)

### Initial Setup

1. **Authenticate with Google Cloud**

```bash
gcloud auth login && gcloud auth application-default login
```

2. **Set your project**

Replace `YOUR_PROJECT_ID` with your actual project ID.

```bash
gcloud config set project YOUR_PROJECT_ID
```

3. **Configure environment**

Edit the `.env` file to set your project ID.

```bash
# Edit .env file
export GOOGLE_CLOUD_PROJECT=your-project-id
```

4. **Initialize infrastructure tracking**

Run the setup script to create the remote state bucket (first time only):

```bash
./scripts/setup_infra.sh
```

### Deploy Resources

```bash
# Validate Terraform configuration
make validate-tf

# Deploy infrastructure
make infrastructure
```

This will create:

- GCS buckets for data storage (raw, staging, processed)
- Pub/Sub topics (if needed)
- IAM roles and service accounts
- Other required GCP resources

### Verification

After deployment, verify resources in the Google Cloud Console:

- **Cloud Storage**: Check for created buckets
- **Pub/Sub**: Verify topics and subscriptions
- **IAM**: Review service account permissions

## Data Model

### Interaction Event

The core data structure for conversation interactions:

```json
{
  "interaction_id": "uuid",
  "conversation_id": "uuid", 
  "timestamp": "iso8601",
  "role": "user|model",
  "content": "string"
}
```

### Processing Pipeline

1. **Raw Storage**: Events stored as JSONL in `gs://bucket/raw/YYYY/MM/DD/HH/`
2. **Staging**: Grouped by conversation in `gs://bucket/staging/YYYY/MM/DD/`
3. **Results**: Security analysis results in `gs://bucket/results/YYYY/MM/DD/`

## Security Analysis

The batch analyzer uses Vertex AI to evaluate conversations for security threats including:

- Prompt injection attacks
- Jailbreak attempts
- Data exfiltration attempts
- Other security vulnerabilities

Analysis uses a configurable prompt template (default: `prompts/security-judge.prompt.yml`) that instructs the LLM to act as a security judge.

## Monitoring and Observability

- **Structured Logging**: All services use `log/slog` for structured logging
- **OpenTelemetry**: Instrumentation support for metrics and traces
- **Error Handling**: Comprehensive error handling and recovery mechanisms

## Project Structure

```ini
.
├── cmd/                          # Service entry points
│   ├── ingestor/                # REST API service
│   ├── loader/                  # Hourly archival job
│   ├── batch-job/               # Daily analysis job
│   └── dataset-loader/          # Dataset ingestion utility
├── internal/
│   ├── app/                     # Application services
│   │   ├── ingestor/           # Ingestor business logic
│   │   ├── loader/             # Loader business logic
│   │   ├── batch/              # Batch analyzer logic
│   │   └── dataset_loader/     # Dataset loading logic
│   └── platform/               # Platform integrations
│       ├── kafka/              # Kafka producer/consumer
│       ├── gcs/                # GCS reader/writer
│       ├── vertex/             # Vertex AI client
│       ├── pinecone/           # Pinecone vector DB
│       └── huggingface/        # HuggingFace client
├── features/                   # BDD feature tests
├── specifications/             # OpenAPI specs
├── prompts/                    # LLM prompt templates
├── terraform/                  # Infrastructure as code
└── scripts/                    # Utility scripts
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run tests: `make test-go`
4. Run linter: `make lint-go`
5. Format code: `make fmt-go`
6. Generate code: `make generate-go`
7. Build: `make build-go`
8. Submit a pull request

## License

[Add your license information here]

## Additional Documentation

- `SPECIFICATION.md` - Detailed system specification
- `PLAN.md` - Implementation plan and architecture decisions
- `TESTING.md` - Comprehensive testing documentation
- `SETUP.md` - GCS log ingestion setup guide
- `GEMINI.md` - Development guidelines

## Support

For issues, questions, or contributions, please open an issue on GitHub.
