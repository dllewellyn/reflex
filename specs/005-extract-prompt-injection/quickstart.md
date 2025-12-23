# Quickstart: Extract Prompt Injection

## Prerequisites

- **Google Cloud Project** with Vertex AI enabled.
- **Pinecone Account** and Index.
- **Service Account** with permissions:
    - `roles/aiplatform.user`
    - `roles/storage.objectViewer`
    - `roles/storage.objectCreator` (for logs/markers if used)
- **Environment Variables**:
    - `GCP_PROJECT_ID`
    - `GCP_LOCATION`
    - `PINECONE_API_KEY`
    - `PINECONE_INDEX_HOST`

## Build

```bash
go build -o bin/extract-injections cmd/extract-injections/main.go
```

## Run

```bash
# Process a specific batch result file
./bin/extract-injections \
  --input "gs://my-bucket/results/2025/01/01/prediction-model-123.jsonl" \
  --dry-run=false
```

## Flags

| Flag | Description | Default |
|---|---|---|
| `--input` | GCS URI pattern for batch results (e.g., `gs://bucket/*.jsonl`) | Required |
| `--prompt` | Path to the extraction prompt file | `prompts/extract-injection.prompt.yml` |
| `--dry-run` | If true, prints extractions but does not upsert | `false` |

## Testing

### Unit Tests
```bash
go test ./internal/app/extract/...
```

### Manual Test
1. Upload a mock result file to GCS: `gs://test-bucket/mock-result.jsonl`
2. Run the command with `--dry-run=true`.
3. Verify output logs show extracted strings.

```