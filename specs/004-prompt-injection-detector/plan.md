# Implementation Plan: Prompt Injection Detector

**Branch**: `004-prompt-injection-detector`
**Input**: Feature specification from `specs/004-prompt-injection-detector/spec.md`

## Summary

Implement a real-time detection service that queries the Pinecone vector database to identify prompt injection attacks. This service relies on the data ingested by `001-ingest-datasets`.

## Dependencies

*   **Data Ingestion**: The Pinecone index must be populated with prompt injection datasets (e.g., `deadbits/vigil-jailbreak` or `deepset/prompt-injections`) using the tooling from `001-ingest-datasets`.

## Phase 1: Detector Service

**Goal**: Real-time analysis endpoint.

1.  **API Scaffolding**:
    *   Create `cmd/detector`.
    *   Implement `POST /analyze`.
2.  **Pinecone Query**:
    *   Receive user prompt.
    *   Send "Query" request to Pinecone Inference API (passing raw text).
    *   Retrieve top `k=1` match with score.
3.  **Decision Logic**:
    *   If `score >= 0.9` -> `ATTACK`.
    *   Else -> `SAFE`.
4.  **Response**:
    *   Return JSON with status and score.

## Phase 2: Integration & Testing

1.  **BDD Tests**:
    *   Create `contracts/detection.feature`.
    *   Scenarios for "Standard Greeting", "Ignore Instructions", "DAN mode".

## Verification

*   **Unit Tests**: Mock Pinecone client.
*   **Integration Tests**: Run against the real Free Tier index with test vectors (requires the index to be pre-populated).