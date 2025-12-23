Feature: Daily LLM Judge Batch Analyzer
  As a security compliance officer
  I want to analyze daily conversations using an LLM
  So that I can detect and alert on high-risk interactions

  Scenario: Process daily conversations
    Given GCS contains raw interaction files for date "2025-12-12"
    And the interactions belong to multiple conversations
    When the Batch Analyzer job is triggered for date "2025-12-12"
    Then it should read all files containing interactions for "2025-12-12"
    And it should group interactions by "conversation_id"
    And it should format each conversation into a Vertex AI Batch Prediction request
    And it should submit the batch job to Vertex AI

  Scenario: Handle conversations spanning hourly files
    Given a conversation "session-123" has interactions in file "10/chunk-1.jsonl" and "11/chunk-1.jsonl"
    When the Batch Analyzer aggregates the data
    Then it should combine all interactions for "session-123" into a single transcript within the batch request

  Scenario: Alert on high-risk findings
    Given Vertex AI returns a batch result containing a high-risk finding
    And the finding has "injection_score" > 0.8
    When the Batch Analyzer processes the results
    Then it should publish a "Security Alert" message to the "security-alerts" Kafka topic
    And the alert should contain the "conversation_id" and the "reasoning"

  Scenario: Ignore low-risk findings
    Given Vertex AI returns a batch result containing a low-risk finding
    And the finding has "injection_score" < 0.5
    When the Batch Analyzer processes the results
    Then it should NOT publish a message to the "security-alerts" Kafka topic
