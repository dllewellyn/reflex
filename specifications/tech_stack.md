Reflex: Technology Stack

Architecture: Centralized Control Plane (Go).
Infrastructure: Google Cloud + Confluent Cloud.

1. The Hub (Control Plane - Go)

High-concurrency stream processing and AI orchestration.

Core Runtime

Language: Go 1.24

Compute: Google Cloud Run (Services & Jobs).

Essential Libraries

Category

Library

Package / Import

Why?

Streaming

Confluent Kafka

github.com/confluentinc/confluent-kafka-go/kafka

High-perf C-binding (librdkafka). Handles backpressure better than native Go clients.

Dependency Injection

Google Wire

github.com/google/wire

Compile-time DI. Generates clean code without runtime reflection magic. Essential for testing.

Math/Stats

Gonum

gonum.org/v1/gonum/stat

Calculates Z-Scores, Mean, and Variance for the "Traffic Physics" engine.

Configuration

Envconfig

github.com/kelseyhightower/envconfig

Loads 12-factor env vars into structs with zero boilerplate.

Logging

Slog

log/slog (Stdlib)

Standard structured JSON logging. No external dependency needed in Go 1.21+.

Telemetry

OpenTelemetry

go.opentelemetry.io/otel

Distributed tracing. Critical to prove "2-minute promotion" latency claims to judges.

Google Cloud SDKs

Service

Package

Purpose

Storage

cloud.google.com/go/storage

Downloading logs, moving files (Rewrite).

Firestore

cloud.google.com/go/firestore

Tenant config lookup and atomic URL updates.

Vertex AI

cloud.google.com/go/aiplatform/apiv1

Calling the AutoML prediction endpoint.

Evaluation

GitHub Models

-

Benchmarking and evaluating different LLMs for the "Judge" role.

Testing & QA

Category

Library

Purpose

BDD Testing

Godog

github.com/cucumber/godog

Unit Testing

Testify

github.com/stretchr/testify

Mocking

Mockery

github.com/vektra/mockery

2. Infrastructure (Terraform)

Infrastructure-as-Code for reproducible deployments.

Provider: hashicorp/google (v5.0+)

Provider: confluentinc/confluent (For Kafka Topics/Clusters).

Modules:

terraform/hub: Deploys Cloud Run, Firestore, Pub/Sub, Vertex AI.

3. Development Tools

Protobuf: For defining the internal schema between Ingestor and Worker.

Buf: (github.com/bufbuild/buf) Modern Protobuf/gRPC toolchain (cleaner than protoc).

Wrk: (github.com/wg/wrk) HTTP benchmarking tool to generate "Fake Viral Traffic" for the demo.