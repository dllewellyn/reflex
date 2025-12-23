Feature: Hourly Loader Job
  As a data engineer
  I want to archive raw interactions from Kafka to GCS
  So that they are preserved for batch analysis

  Scenario: Archive messages to GCS
    Given the configured Kafka topic contains 100 messages
    And the GCS bucket "security-data-lake" is accessible
    When the Loader job is triggered
    Then it should consume all 100 messages
    And it should write a single JSONL file to "gs://security-data-lake/raw/" with the current timestamp prefix
    And the GCS object key should follow the pattern "raw/<conversation_id>/YYYY/MM/DD/HH/chunk-<uuid>.jsonl"
    And it should commit the Kafka offsets after the write is successful

  Scenario: Handle zero messages
    Given the configured Kafka topic is empty
    When the Loader job is triggered
    Then it should exit gracefully without writing any file
    And it should not report an error

  Scenario: Ensure data consistency on failure
    Given the configured Kafka topic contains messages
    And the GCS service is down
    When the Loader job is triggered
    Then it should fail to write to GCS
    And it should NOT commit any Kafka offsets
    And the process should exit with a non-zero status code
