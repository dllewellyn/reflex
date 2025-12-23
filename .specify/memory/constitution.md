<!--
Sync Impact Report:
- Version change: 1.7.0 → 1.8.0
- Modified Principles:
  - IX. Observability & OpenTelemetry
- Templates requiring updates:
  - ✅ .specify/templates/plan-template.md
  - ✅ .specify/templates/spec-template.md
  - ✅ .specify/templates/tasks-template.md
-->
# Reflex Reflex Constitution

## Core Principles

### I. BDD Test Cases in Plan Phase
BDD test cases must be generated as part of the `plan` phase. This is a non-negotiable prerequisite to proceeding with implementation.

### II. Adherence to Directory Structure
The project must strictly follow the directory structure defined in `specifications/directory_structure.md` to ensure consistency and predictability across the codebase.

### III. Core Architecture
The system will process GCS access logs to identify high-traffic "viral" requests. These requests will trigger an upgrade from standard GCS storage to a Firebase CDN. This logic is orchestrated via a Kafka stream, with an edge function in `functions/` handling the CDN upgrade.

### IV. Primary Language: Golang
All backend services and components shall be written in Golang. The only exception is the Firebase edge function, which is implemented in Node.js/TypeScript as required by its runtime.

### V. Abstraction with Interfaces
Use interfaces to abstract concrete implementations of external services (e.g., Kafka, Firestore, GCS). This promotes modularity, testability, and allows for easier replacement of components.

### VI. Dependency Injection with Wire
The `wire` library must be used for managing dependency injection in all Golang services to ensure a clean and maintainable application structure.

### VII. Clean Coding Practices
All code must adhere to established clean coding best practices, emphasizing readability, maintainability, and simplicity (YAGNI).

### VIII. Secrets Management & Environment
Local development must use valid `.env` files derived from `.env.example` for non-sensitive configuration. Sensitive secrets must be managed using an appropriate secrets manager (e.g., Google Secret Manager). For CI/CD, secrets must be securely injected via GitHub Actions secrets. Hardcoding secrets is strictly prohibited.

### IX. Observability & OpenTelemetry
Logging must be implemented as a core part of the application logic using OpenTelemetry. Standard library loggers (e.g., `log`, `slog`) should be bridged to or replaced by OpenTelemetry loggers to ensure correlation between logs, traces, and metrics. All services must utilize OpenTelemetry for tracing, metrics, and logging to ensure comprehensive system observability.

### X. Continuous Verification
Every task implementation must conclude with a successful build. Additionally, final verification must confirm that GitHub Actions workflows execute successfully using the GitHub CLI (e.g., `gh run list/view`) to ensure no broken states are committed.

### XI. Realistic End-to-End Testing
Feature tests (E2E) must be capable of executing against "real" infrastructure (e.g., actual Kafka instances, Firebase projects) rather than solely relying on mocks. These tests must be fully automated and executable within the GitHub Actions CI environment.

### XII. Serverless & Scale-to-Zero First
Infrastructure choices must prioritize "serverless" and "scale-to-zero" components (e.g., Cloud Run, Cloud Functions, Firestore). Use of always-on infrastructure (e.g., GKE, VMs) is permitted **only** if serverless options are technically infeasible, and this decision must be explicitly justified in the Implementation Plan.

### XIII. Infrastructure Planning
The planning phase must explicitly evaluate if new infrastructure is required to deliver the feature. This evaluation, including the selection of components and compliance with the serverless preference, must be documented in the Implementation Plan.

### XIV. Schema-Driven Development & Code Generation
All external data interfaces (e.g., Kafka topics, GCS objects, Pinecone records) must be defined as JSON Schemas in the `specifications/schemas/` directory. API interfaces must be defined using OpenAPI (e.g., `specifications/openapi.yaml`). All corresponding code (structs, clients, servers) must be automatically generated from these definitions using appropriate tools. Manual maintenance of these structures in the codebase is prohibited to ensure a single source of truth.

## Governance

All code reviews must explicitly verify compliance with the principles outlined in this constitution. Any deviation requires explicit justification and approval.

**Version**: 1.8.0 | **Ratified**: 2025-12-09 | **Last Amended**: 2025-12-15