Feature: Ingestor Service
  As a security analyst
  I want to reliably ingest user-model interactions
  So that I can analyze them for potential security threats

  Scenario: Successfully ingest a valid interaction
    Given the Ingestor service is running
    And the configured Kafka topic is available
    When I send a POST request to "/ingest" with the following JSON:
      """
      {
        "interaction_id": "550e8400-e29b-41d4-a716-446655440000",
        "conversation_id": "c6168400-e29b-41d4-a716-446655440000",
        "timestamp": "2025-12-12T10:00:00Z",
        "role": "user",
        "content": "Hello, world!"
      }
      """
    Then the response status code should be 202
    And the message should be published to the configured Kafka topic

  Scenario: Reject invalid JSON payload
    Given the Ingestor service is running
    When I send a POST request to "/ingest" with the following JSON:
      """
      {
        "interaction_id": "invalid-uuid",
        "role": "unknown"
      }
      """
    Then the response status code should be 400
    And the response body should contain an error message about schema validation

  Scenario: Handle Kafka unavailability
    Given the Ingestor service is running
    And the Kafka cluster is unreachable
    When I send a POST request to "/ingest" with a valid interaction payload
    Then the response status code should be 500
    And the response body should indicate an internal server error
