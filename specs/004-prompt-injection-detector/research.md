# Research: Prompt Injection Detector

## 1. Vector Database: Pinecone

We will use **Pinecone** (Free Tier) for vector storage and similarity search.

*   **Plan**: Starter (Free)
*   **Feature**: **Pinecone Inference API**.
    *   We will leverage the integrated inference capability to avoid managing a separate embedding model service.
    *   Pinecone handles text-to-vector conversion query.
*   **SDK**: Official Go SDK: `github.com/pinecone-io/go-pinecone`.

## 2. Dataset Strategy

This feature relies on the data ingested by **`001-ingest-datasets`**.

*   **Primary Dataset**: `deadbits/vigil-jailbreak` (or `deepset/prompt-injections`).
*   **Mechanism**: The detector queries the index where `001-ingest-datasets` has stored the attack signatures.
*   **Matching Logic**: We use Cosine Similarity to find if a user's prompt is semantically close to any known attack in the dataset.

## 3. Infrastructure

*   **Resource**: Pinecone Index (Serverless).
*   **Metric**: Cosine Similarity.