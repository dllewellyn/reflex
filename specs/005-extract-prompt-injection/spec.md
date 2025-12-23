# Feature Specification: Extract Prompt Injection Strings

**Feature Branch**: `005-extract-prompt-injection`  
**Created**: 2025-12-17  
**Status**: Draft  
**Input**: User description: "Given the results of the batch AI results, we need a command to extract all of those where prompt_injection is true, pass it to an LLM to extract a list of "prompt injection" strings from it, and then upsert those into the pinecone database"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Extract and Upsert Known Attacks (Priority: P1)

As a Security Researcher, I want to process the results of our batch prompt injection analysis to automatically extract and catalog the specific injection strings used, so that I can improve our vector database of known threats without manual labeling.

**Why this priority**: Automates the feedback loop between detection (Batch Judge) and prevention (Vector Search), making the system smarter over time.

**Independent Test**:
1. Create a mock Batch Result file containing positive and negative injection detections.
2. Run the command pointing to this file.
3. Verify that the system calls the extraction LLM only for the positive cases.
4. Verify that the extracted strings are upserted into the expected Pinecone index.

**Acceptance Scenarios**:

1. **Given** a GCS path containing Vertex AI Batch Prediction results (JSONL), **When** the command is run, **Then** it iterates through all prediction records.
2. **Given** a record where the judge output indicates `prompt_injection: true`, **When** processed, **Then** the original conversation transcript is sent to an LLM with instructions to extract the specific injection payload(s).
3. **Given** a list of extracted injection strings from the LLM, **When** upserting, **Then** they are converted to vectors and stored in Pinecone with metadata `source=auto-extracted` and `label=injection`.
4. **Given** a record where `prompt_injection: false`, **When** processed, **Then** it is skipped to save costs.

---

### User Story 2 - Idempotency and Tracking (Priority: P2)

As an Operator, I want the extraction process to skip already-processed results, so that I can re-run the command on the same bucket without paying for duplicate LLM extraction calls or creating duplicate database entries.

**Why this priority**: Prevents waste of tokens and ensures database hygiene.

**Independent Test**: Run the command twice on the same input file; the second run should report 0 extractions/upserts.

**Acceptance Scenarios**:

1. **Given** a batch result file that has been partially processed, **When** the command is re-run, **Then** it identifies already processed interaction IDs (e.g., via a tracking file or checking Pinecone) and skips them.
2. **Given** duplicate injection strings extracted (e.g., same attack used in multiple sessions), **When** upserting to Pinecone, **Then** the vectors are deduplicated or the upsert operation handles it gracefully (idempotent write).

### Edge Cases

- **Extraction Failure**: The LLM fails to identify a specific string (returns "None"). The system should log this and not upsert anything.
- **Malformed Batch Output**: The input JSONL line is corrupt. The system should log the error and continue to the next line.
- **Pinecone Unavailability**: The vector DB is down. The command should retry or exit with a clear error state, allowing resume later.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a CLI command (e.g., `extract-injections`) that accepts a GCS URI pattern for batch results.
- **FR-002**: System MUST parse Vertex AI Batch Prediction output files to identify positive `prompt_injection` findings.
- **FR-003**: System MUST define and use a new Prompt Template (e.g., `prompts/extract-injection.prompt.yml`) to instruct an LLM to extract the exact injection string(s) from a conversation transcript.
- **FR-004**: System MUST invoke the LLM (Vertex AI) for each positive finding using the extraction prompt.
- **FR-005**: System MUST embed the extracted strings using the configured embedding model (consistent with existing Pinecone setup).
- **FR-006**: System MUST upsert the resulting vectors to the Pinecone index with metadata (including `interaction_id`, `extracted_at`, `type: auto-extracted`).
- **FR-007**: System MUST support dry-run mode to preview what would be extracted/upserted without making changes.

### Key Entities *(include if feature involves data)*

- **BatchResult**: The input record from GCS containing the Judge's decision and the original transcript.
- **ExtractionPrompt**: The instruction set for the LLM to isolate the attack vector.
- **ExtractedAttack**: The output string(s) identified as the injection attempt.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Command can process a batch result file of 10,000 records in under 30 minutes (mostly dependent on LLM concurrency).
- **SC-002**: >95% of records marked `prompt_injection: true` result in at least one extracted string (unless false positive).
- **SC-003**: Injected vectors are searchable in Pinecone immediately after upsert.
- **SC-004**: System handles API rate limits (Vertex AI, Pinecone) with automatic backoff/retry.

### Assumptions

- The Batch AI results contain the original input (transcript) or a reference that allows retrieval.
- A standard embedding model (e.g., `text-embedding-004`) is already configured and available for use.
- The Pinecone index schema allows adding new records with arbitrary metadata.