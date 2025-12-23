# Makefile

# Variables
GO_CMD=go
GOPATH=$(shell $(GO_CMD) env GOPATH)
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_LINT=$(GOPATH)/bin/golangci-lint run
GO_FMT=gofmt
DOCKER=docker
TIDY=$(GO_CMD) mod tidy

TERRAFORM=terraform

# Project paths
LOADER_DIR=cmd/loader
INGESTOR_DIR=cmd/ingestor
BATCH_DIR=cmd/batch-job
DATASET_LOADER_DIR=cmd/dataset-loader
EVALUATE_DIR=cmd/evaluate
EXTRACT_DIR=cmd/extract-injections
HUB_TF_DIR=terraform

.PHONY: all build test lint run-ingestor docker-build infrastructure validate-tf clean init fmt-go fmt-check generate-wire tools

all: test build

# --- Aggregate Targets ---

build: build-go

test: test-go

lint: lint-go

# --- Initialization ---

init:
	@echo "Initializing Go dependencies..."
	$(GO_CMD) mod download
	@echo "Initializing Terraform..."
	cd $(HUB_TF_DIR) && $(TERRAFORM) init


tools:
	@echo "Installing tools..."
	$(GO_CMD) install github.com/atombender/go-jsonschema/...@latest
	$(GO_CMD) install github.com/google/wire/cmd/wire@latest
	$(GO_CMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO_CMD) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# --- Go Targets ---

generate-schemas:
	@echo "Generating schemas..."
	./scripts/generate_schemas.sh

generate-wire:
	@echo "Generating Wire code..."
	PATH="$(GOPATH)/bin:$(PATH)" $(GO_CMD) generate -tags wireinject ./...

generate-api:
	@echo "Generating API server..."
	mkdir -p internal/app/ingestor/server
	$(GOPATH)/bin/oapi-codegen -package server -generate types,chi-server,spec -o internal/app/ingestor/server/server.gen.go specifications/ingestor-api.yaml

generate-go: generate-schemas generate-wire generate-api
	@echo "Go code generation complete."

generate: generate-go
	@echo "Generating all code..."
	
build-go: generate-go tidy build-ingestor build-loader build-batch build-dataset-loader build-extract build-evaluate

build-ingestor:
	@echo "Building Ingestor..."
	$(GO_BUILD) -o ./bin/ingestor ./$(INGESTOR_DIR)

build-loader:
	@echo "Building Loader..."
	$(GO_BUILD) -o ./bin/loader ./$(LOADER_DIR)

build-batch:
	@echo "Building Batch Job..."
	$(GO_BUILD) -o ./bin/batch-job ./$(BATCH_DIR)

build-dataset-loader:
	@echo "Building Dataset Loader..."
	$(GO_BUILD) -o ./bin/dataset-loader ./$(DATASET_LOADER_DIR)

build-extract:
	@echo "Building Extract Injections..."
	$(GO_BUILD) -o ./bin/extract-injections ./$(EXTRACT_DIR)

build-evaluate:
	@echo "Building Evaluate..."
	$(GO_BUILD) -o ./bin/evaluate ./$(EVALUATE_DIR)

test-go:
	@echo "Running Go Tests..."
	$(GO_TEST) ./...
	$(GO_TEST) -tags=integration ./...

fmt-go:
	@echo "Formatting Go code..."
	$(GO_FMT) -w .

fmt-check:
	@echo "Checking Go code formatting..."
	@if [ -n "$$($(GO_FMT) -l .)" ]; then \
		echo "The following files need formatting:"; \
		$(GO_FMT) -l .; \
		echo "Run 'make fmt-go' to fix formatting issues."; \
		exit 1; \
	fi
	@echo "All Go files are properly formatted."

lint-go:
	@echo "Linting Go..."
	$(GO_LINT) --timeout=5m ./...

run-ingestor:
	@echo "Running Ingestor..."
	$(GO_CMD) run $(INGESTOR_DIR)/main.go

docker-build: docker-build-ingestor

docker-build-ingestor:
	@echo "Building Ingestor Docker Image..."
	$(DOCKER) build -f $(INGESTOR_DIR)/Dockerfile -t ingestor:latest .

# --- Terraform Targets ---

validate-tf:
	@echo "Validating Hub Terraform..."
	cd $(HUB_TF_DIR) && $(TERRAFORM) init -backend=false && $(TERRAFORM) validate

infrastructure: prepare-function-source
	@echo "Deploying Hub Resources..."
	cd $(HUB_TF_DIR) && $(TERRAFORM) init && $(TERRAFORM) apply -auto-approve

prepare-function-source:
	@echo "Preparing Cloud Function source..."
	rm -rf bin/function-source
	mkdir -p bin/function-source
	cp go.mod go.sum bin/function-source/
	cp -r internal bin/function-source/
	cp cmd/batch-result-trigger/*.go bin/function-source/


# --- Clean ---
clean:
	rm -rf bin

tidy:
	$(GO_CMD) mod tidy
