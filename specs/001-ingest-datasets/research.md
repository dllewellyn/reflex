# Research: Ingest Prompt Injection Datasets

**Feature**: Ingest Prompt Injection Datasets (001-ingest-datasets)
**Status**: Resolved

## 1. CLI vs Job Configuration

**Decision**: Use `kelseyhightower/envconfig` for configuration.
**Rationale**: The project already uses `envconfig`. The component is a batch job (Cloud Run Job), which is best configured via environment variables. Complex CLI flags (Cobra) are unnecessary complexity for a non-interactive tool.
**Alternatives**: `spf13/cobra` (rejected: overhead), standard `flag` (rejected: less robust for struct mapping).

## 2. HuggingFace Dataset Access

**Decision**: Fetch Parquet files directly via HTTP and parse with `github.com/parquet-go/parquet-go`.
**Rationale**: HuggingFace datasets are typically exposed as Parquet files. The `parquet-go` library is already in `go.mod`. This avoids the need for a Python bridge or a complex/unmaintained Go wrapper for the HF Hub API.
**Alternatives**: Python script (rejected: constitution requires Go), `huggingface-go` unofficial libs (rejected: maturity concerns).

## 3. Pinecone Integration

**Decision**: Use `github.com/pinecone-io/go-pinecone/v4`.
**Rationale**: It is the official SDK and is already present in `go.mod` (indirectly, likely via other dependencies or ready to be promoted).
**Alternatives**: REST API directly (rejected: SDK provides type safety and connection pooling).

## 4. Vector Embedding Generation

**Decision**: Use Pinecone's "Inference API" or "integrated embedding model" if the index is configured that way (as per Spec FR-003 "using the system's configured embedding model").
**Rationale**: The spec assumes the index is pre-provisioned with an embedding model. This simplifies the client logic (just send text).
**Clarification**: If the index does *not* support server-side embedding, we would need a local model or an API call (e.g., Vertex AI). The Plan assumes Pinecone handles it or we use the `embedding` model configured in Pinecone. The spec FR-003 says "using the system's configured embedding model" but FR-004 says "upsert vectors". If we upsert vectors, we must generate them first.
**Refinement**: If Pinecone index is "serverless" with "integrated inference", we upsert *records* with text, and it generates embeddings. If it's a standard index, we upsert *vectors*.
**Assumption**: Given the spec mentions "upsert vectors" (FR-004) AND "generate vector embeddings" (FR-003), the responsibility lies with *this system*.
**Correction**: I will assume we need to generate embeddings. However, looking at the tools available to me (`create-index-for-model`), it seems Pinecone can handle embedding.
**Final Decision**: Check if the intention is to use Pinecone's inference. The spec FR-003 says "System MUST generate vector embeddings... using the system's configured embedding model". This could mean "call the model". For now, I will assume we need an embedding service interface (which could be Vertex AI or Pinecone Inference).

## 5. Idempotency Strategy

**Decision**: Use deterministic IDs for vectors based on the content hash (e.g., SHA256 of the prompt text).
**Rationale**: This ensures that re-ingesting the same text results in the same ID, causing an update (upsert) rather than a duplicate.