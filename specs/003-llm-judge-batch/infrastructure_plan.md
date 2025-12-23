# Infrastructure Plan: LLM Judge Batch

This document outlines the infrastructure required to deploy the "LLM Judge" scheduled batch pipeline.

## 1. Google Cloud Resources (Terraform)

The following resources need to be defined in `terraform/`:

### A. Service Accounts
1.  **Loader Service Account** (`loader-sa`)
    *   **Purpose**: Identity for the hourly job that moves data from Kafka to GCS.
    *   **Permissions**:
        *   `roles/storage.objectAdmin` (Read/Write to GCS bucket)
        *   *Kafka Access*: Depending on Kafka hosting (e.g., Secret Manager accessor if SASL/SSL credentials are needed).

2.  **Batch Judge Service Account** (`batch-judge-sa`)
    *   **Purpose**: Identity for the daily job that triggers Vertex AI.
    *   **Permissions**:
        *   `roles/storage.objectAdmin` (Read raw logs, write batch results)
        *   `roles/aiplatform.user` (Create/Manage Vertex AI Batch Jobs)
        *   `roles/logging.logWriter`

### B. Cloud Run Jobs
Two distinct Cloud Run Jobs will be deployed.

#### Job 1: `interaction-loader`
*   **Image**: `us-central1-docker.pkg.dev/<project>/flash-cache-repo/loader:latest`
*   **Service Account**: `loader-sa`
*   **Environment Variables**:
    *   `GCP_PROJECT_ID`: (Project ID)
    *   `GCS_BUCKET_NAME`: Name of the conversation history bucket.
    *   `KAFKA_BROKERS`: List of Kafka brokers.
    *   `KAFKA_TOPIC`: `raw-interactions`
    *   `KAFKA_CONSUMER_GROUP`: `gcs-loader-v1`
*   **Resources**: 1 CPU, 512MB RAM.

#### Job 2: `llm-judge-trigger`
*   **Image**: `us-central1-docker.pkg.dev/<project>/flash-cache-repo/batch-judge:latest`
*   **Service Account**: `batch-judge-sa`
*   **Environment Variables**:
    *   `GCP_PROJECT`: (Project ID)
    *   `GCP_LOCATION`: `us-central1`
    *   `INPUT_BUCKET`: (Same as `GCS_BUCKET_NAME`)
    *   `OUTPUT_BUCKET`: (Same as `GCS_BUCKET_NAME`)
    *   `MODEL_ID`: `publishers/google/models/gemini-2.5-flash`
    *   `PROMPT_PATH`: `prompts/security-judge.prompt.yml` (Note: This file must be baked into the image).

### C. Cloud Scheduler
Two scheduler jobs to manage the cadence.

1.  **`trigger-loader-hourly`**
    *   **Schedule**: `0 * * * *` (Every hour).
    *   **Target**: Cloud Run Job `interaction-loader`.

2.  **`trigger-judge-daily`**
    *   **Schedule**: `0 2 * * *` (2:00 AM Daily).
    *   **Target**: Cloud Run Job `llm-judge-trigger`.

## 2. Application Updates

### A. `cmd/batch-job` Refactoring
The current `cmd/batch-job/main.go` relies heavily on CLI flags (`flag.String`). To work seamlessly with Cloud Run Jobs, it should be updated to prioritize Environment Variables.

**Recommendation**:
Update `main.go` to use `os.Getenv` for defaults if flags are not provided, or use a library like `github.com/kelseyhightower/envconfig`.

| Env Var | Corresponding Flag |
|---|---|
| `GCP_PROJECT` | `--project` |
| `GCP_LOCATION` | `--location` |
| `GCS_INPUT_URI` | `--input` (Note: Logic might be needed to auto-calculate the "yesterday" path if this is omitted in a cron context) |
| `GCS_OUTPUT_PREFIX` | `--output` |

**Path Calculation Logic**:
The Daily Batch Job needs to know *what* to process. The application code should likely support a "dynamic mode" where if no input URI is provided, it calculates the path for "Yesterday" (e.g., `gs://<bucket>/raw/*/<yesterday_date>/*.jsonl`).

### B. Dockerfiles

**`build/Dockerfile.loader`**
```dockerfile
FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go build -o loader cmd/loader/main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/loader /
CMD ["/loader"]
```

**`build/Dockerfile.batch-job`**
```dockerfile
FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go build -o batch-job cmd/batch-job/main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/batch-job /
COPY prompts/ /prompts/
CMD ["/batch-job"]
```
*Note: We must copy the `prompts/` directory so the application can read the YAML file.*

## 3. Deployment Workflow

1.  **Terraform Apply**: Create Service Accounts and Jobs.
2.  **Build Images**:
    *   `docker build -f build/Dockerfile.loader -t ...`
    *   `docker build -f build/Dockerfile.batch-job -t ...`
3.  **Push to Artifact Registry**.
4.  **Update Jobs**: Ensure Cloud Run Jobs point to the new image digest.
