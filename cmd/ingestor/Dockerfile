# Use official Golang image to build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (required by confluent-kafka-go)
# Added bash for schema generation script
RUN apk add --no-cache build-base bash

# Copy modules manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate schemas
RUN chmod +x ./scripts/generate_schemas.sh && ./scripts/generate_schemas.sh

# Build
RUN go build -tags musl -o /ingestor ./cmd/ingestor

# Runtime image
FROM alpine:latest

# Install runtime dependencies for Kafka (librdkafka)
# Note: confluent-kafka-go static build (-tags musl) usually bundles it,
# but sometimes library dependencies are needed.
# For simplicity with musl build, static is preferred.

WORKDIR /root/

COPY --from=builder /ingestor .

EXPOSE 8080

CMD ["./ingestor"]
