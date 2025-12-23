# Feature Specification: Ingest Prompt Injection Datasets

**Feature Branch**: `001-ingest-datasets`  
**Created**: 2025-12-15  
**Status**: Draft  
**Input**: User description: "Ingest existing prompt injection datasets from huggingface and load into pinecone"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ingest Standard Injection Datasets (Priority: P1)

As a System Administrator, I want to trigger an ingestion job that pulls known prompt injection datasets from HuggingFace and stores them in Pinecone, so that the system can use this data for similarity-based attack detection.

**Why this priority**: This is the core functionality required to populate the vector database with threat data. Without this, the system has no knowledge base for detection.

**Independent Test**: Can be tested by running the ingestion command with a small subset of the dataset and verifying the vectors appear in the target Pinecone index.

**Acceptance Scenarios**:

1. **Given** a valid HuggingFace dataset configuration (e.g., `deepset/prompt-injections`), **When** the ingestion job is triggered, **Then** the system downloads the dataset rows.
2. **Given** the downloaded text data, **When** processing, **Then** the system generates vector embeddings for each text entry.
3. **Given** generated vectors, **When** uploading, **Then** the vectors are stored in the configured Pinecone index with metadata including the original text and label (if available).
4. **Given** the ingestion completes, **When** I query Pinecone for a known phrase from the dataset, **Then** I receive a high-similarity match.

---

### User Story 2 - Idempotent Re-ingestion (Priority: P2)

As an Administrator, I want the ingestion process to be idempotent, so that running the job multiple times does not create duplicate vectors or corrupt the index.

**Why this priority**: Ensures data integrity and allows for safe retries in case of partial failures.

**Independent Test**: Run the ingestion job twice on the same dataset and verify the count of vectors in Pinecone remains constant (or updates existing ones) rather than doubling.

**Acceptance Scenarios**:

1. **Given** a dataset has already been ingested, **When** I trigger the ingestion job again, **Then** the system updates existing vectors or skips them, ensuring no duplicates are created.
2. **Given** a partial failure in a previous run, **When** re-run, **Then** the system picks up missing items or overwrites to reach a consistent state.

### Edge Cases

- **HF API Failure**: If HuggingFace is down or rate-limiting, the system should retry with backoff or fail gracefully with a descriptive error.
- **Empty Dataset**: If a configured dataset is empty or malformed, the system should log a warning and proceed or exit based on configuration.
- **Pinecone Limits**: If Pinecone index is full or rate-limited, the system should handle the error and report the failure status.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST be able to connect to the HuggingFace Hub API to fetch datasets.
- **FR-002**: System MUST support configuration of target datasets (e.g., dataset ID, split).
- **FR-003**: System MUST generate vector embeddings for the text content of the dataset rows using the system's configured embedding model.
- **FR-004**: System MUST upsert vectors to a configured Pinecone index.
- **FR-005**: System MUST store original text and classification labels (e.g., 'injection', 'safe') as metadata in Pinecone.
- **FR-006**: System MUST handle large datasets by batching requests to both HuggingFace (if applicable) and Pinecone to avoid memory/rate issues.
- **FR-007**: System MUST log progress (e.g., "Processed 100/1000 records") and any errors encountered during the process.

### Key Entities

- **Dataset Source**: Represents the HuggingFace dataset configuration (ID, name, column mapping).
- **Injection Record**: A single row of data containing the prompt text and its label.
- **Vector Entry**: The processed record containing the embedding vector and associated metadata ready for Pinecone.

### Assumptions

- The target Pinecone index is pre-provisioned or the system has permissions to create it, configured to use an integrated embedding model.
- The dataset on HuggingFace is public or appropriate credentials are provided.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Ingestion of a standard benchmark dataset (e.g., 10,000 records) completes in under 15 minutes (assuming standard network/API speeds).
- **SC-002**: 100% of valid records in the source dataset are represented in the target vector store after a successful run.
- **SC-003**: Querying the vector index with a sample injection prompt returns the corresponding ingested record within the top 5 results.
- **SC-004**: System handles external API rate limits (Source or Destination) without crashing, resulting in a successful completion state.