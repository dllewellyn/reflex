# Quickstart: Process Batch Results

## Prerequisites

- Terraform installed.
- Go 1.23+ installed.
- Access to GCP Project with GCS and Cloud Functions enabled.
- Kafka cluster accessible.

## Deployment

1. **Deploy Infrastructure**:
   ```bash
   cd terraform
   terraform apply
   ```
   This will create the GCS bucket (if not exists) and the Cloud Function trigger.

2. **Verify Trigger**:
   Upload a test JSONL file to the `processed_prompts` bucket:
   ```bash
   gsutil cp tests/fixtures/sample_batch_result.jsonl gs://<YOUR_PROCESSED_BUCKET>/results/test.jsonl
   ```

3. **Verify Kafka Output**:
   Consume from the topic to see the messages:
   ```bash
   # Example using kcat
   kcat -b <BROKER> -t batch-job-results -C
   ```

## Running Locally

To run the `extract-injections` consumer locally:

```bash
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=batch-job-results
go run cmd/extract-injections/main.go
```
