# Quickstart: LLM Judge Batch Job

## Prerequisites
*   Go 1.23+
*   Docker (for GCS/Kafka emulators)
*   `make`

## Running Locally

1.  **Start Infrastructure**:
    ```bash
    make infra-up
    ```
    This starts Kafka and the GCS emulator.

2.  **Seed Data**:
    Use the `ingestor` and `loader` to generate some raw data, or manually upload a file to the local GCS bucket.
    ```bash
    # Example: Upload dummy data
    curl -X POST --data-binary @testdata/raw_chunk.jsonl http://localhost:4443/upload/storage/v1/b/my-bucket/o?name=raw/conv-123/2025/12/14/00/chunk-1.jsonl
    ```

3.  **Run the Batch Job**:
    ```bash
    # Run the batch job with required flags
    go run cmd/batch-job/main.go \
        --project=my-gcp-project \
        --location=us-central1 \
        --input=gs://my-bucket/raw/conv-123/2025/12/14/00/chunk-1.jsonl \
        --output=gs://my-bucket/results/conv-123/2025/12/14/00/output-1.jsonl \
        --model=chat-bison \
        --prompt="Summarize the conversation"
    ```

4.  **Verify Results**:
    Check the Kafka topic for alerts:
    ```bash
    kcat -C -b localhost:9092 -t security-alerts
    ```
