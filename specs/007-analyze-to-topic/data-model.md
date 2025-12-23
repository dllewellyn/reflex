# Data Model: Real-time Analysis

**Feature**: `007-analyze-to-topic`

## 1. Entities

### Interaction Event
*   **Source**: `specifications/schemas/interaction-event.schema.json`
*   **Description**: Represents a single turn of conversation between a user and the model.

| Field | Type | Required | Description |
|---|---|---|---|
| `interaction_id` | UUID | Yes | Unique ID for the event |
| `conversation_id` | UUID | Yes | Session ID |
| `timestamp` | DateTime | Yes | UTC timestamp |
| `role` | Enum | Yes | `user` or `model` |
| `content` | String | Yes | The text payload |

## 2. Topic Schema
*   **Topic Name**: `analysis-requests`
*   **Key**: `conversation_id` (to ensure ordering per conversation)
*   **Value**: JSON-serialized `Interaction Event`
