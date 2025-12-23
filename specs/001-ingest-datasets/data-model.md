# Data Model: Ingest Prompt Injection Datasets

## Entities

### 1. Source Record (HuggingFace)

Represents a single row from the Parquet file downloaded from HuggingFace.

| Field | Type | Description |
|---|---|---|
| `text` | String | The potential injection prompt or safe text. |
| `label` | Int/String | Classification (0=safe, 1=injection, or similar). |
| `split` | String | The dataset split (train/test/validation). |

### 2. Ingestion Record (Internal)

Normalized record processed by the service.

| Field | Type | Description |
|---|---|---|
| `ID` | String | Deterministic hash (SHA256) of `Text`. |
| `Text` | String | The original content. |
| `IsInjection` | Boolean | Normalized label (True if injection). |
| `Source` | String | Name of the source dataset (e.g., "deepset/prompt-injections"). |

### 3. Pinecone Record (Sink)

The entry stored in the Vector Database.

| Field | Type | Description |
|---|---|---|
| `id` | String | Matches `IngestionRecord.ID`. |
| `values` | []Float32 | Vector embedding of `Text`. |
| `metadata` | Map | Key-value pairs. |

**Metadata Schema**:
```json
{
  "text": "string",
  "is_injection": "boolean",
  "source": "string",
  "ingested_at": "timestamp"
}
```