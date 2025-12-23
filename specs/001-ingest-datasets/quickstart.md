# Quickstart: Ingest Prompt Injection Datasets

## Prerequisites

- Go 1.24+
- Pinecone Account & API Key
- HuggingFace Account (for Token, if accessing private datasets)
- A Pinecone Index created (serverless or pod-based)

## Setup

1. **Clone the repository** (if not already done).

2. **Configure Environment**:
   Copy `.env.example` to `.env` and fill in the required values.

   ```bash
   cp .env.example .env
   ```

   **Required Variables**:
   
   | Variable | Description | Example |
   |---|---|---|
   | `PINECONE_API_KEY` | Your Pinecone API Key | `pc_...` |
   | `HF_TOKEN` | HuggingFace Token (optional for public datasets) | `hf_...` |
   | `PINECONE_INDEX_HOST` | The host URL of your Pinecone Index | `https://index-name-xyz.svc.pinecone.io` |

   **Optional Configuration (Defaults exist)**:

   | Variable | Description | Default |
   |---|---|---|
   | `HF_DATASET_ID` | Dataset to ingest | `deepset/prompt-injections` |
   | `HF_SPLIT` | Dataset split | `train` |
   | `HF_TEXT_COL` | Text column name | `text` |
   | `HF_LABEL_COL` | Label column name | `label` |

## Running the Job

Execute the ingestion job using `go run`:

```bash
go run cmd/dataset-loader/main.go
```

## verifying Success

1. Check the logs for "Successfully retrieved..." and "Upserted X records".
2. Log in to the Pinecone Console.
3. Verify the index has vectors and the metadata contains the text.