# Feature Specification: Prompt Injection Detector

**Feature Branch**: `004-prompt-injection-detector`
**Status**: Draft
**Created**: December 14, 2025

## 1. Overview

The **Prompt Injection Detector** provides real-time analysis of user prompts to identify potential injection attacks. It utilizes **Pinecone's Serverless Inference API** to perform semantic similarity searches against a knowledge base of known attack vectors (ingested via `001-ingest-datasets`).

## 2. Requirements

### Functional

*   **FR-001**: The system MUST expose an HTTP endpoint (e.g., `POST /analyze`) that accepts a JSON payload containing the user's prompt.
*   **FR-002**: The system MUST use **Pinecone Integrated Inference** (or the configured embedding model) to automatically generate embeddings for the incoming prompt during the query.
*   **FR-003**: The system MUST query the existing **Pinecone Vector Database** to find the nearest neighbors within the ingested attack dataset.
*   **FR-004**: The system MUST determine if a prompt is an attack based on a similarity threshold.
    *   **Rule**: If the similarity score of the closest match is **>= 0.9**, the prompt is classified as an **ATTACK**.
*   **FR-005**: The response MUST include the classification (`SAFE` or `ATTACK`) and the similarity score.

### Non-Functional

*   **NFR-001**: The service should be designed for low latency (< 200ms) to support real-time guarding.
*   **NFR-002**: Vector database credentials (`PINECONE_API_KEY`) and index host must be managed via environment variables.
*   **NFR-003**: The system MUST utilize the Pinecone Free Tier limits effectively (1 index, serverless).

## 3. Success Criteria

*   **SC-001**: Correctly identifies 100% of exact matches from the attack dataset.
*   **SC-002**: Identifies semantically similar variants of attacks with a score >= 0.9.
*   **SC-003**: Benign prompts result in similarity scores < 0.9.