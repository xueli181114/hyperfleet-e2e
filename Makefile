# HyperFleet E2E Test Makefile

# Variables
BINARY_NAME=e2e-runner
BUILD_DIR=build
FEATURES_DIR=features
TEST_RESULTS_DIR=test-results

# Go build settings
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
VERSION=$(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Default target
.PHONY: all
all: clean build test

# Build the test runner
.PHONY: build
build:
	@echo "Building e2e-runner..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/e2e-runner

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Run tests with Ginkgo (default)
.PHONY: test
test: build
	@echo "Running E2E tests with Ginkgo..."
	./$(BUILD_DIR)/$(BINARY_NAME) --mode ginkgo

# Run tests with Godog
.PHONY: test-godog
test-godog: build
	@echo "Running E2E tests with Godog..."
	./$(BUILD_DIR)/$(BINARY_NAME) --mode godog

# Run both Ginkgo and Godog tests
.PHONY: test-all
test-all: build
	@echo "Running E2E tests with both Ginkgo and Godog..."
	./$(BUILD_DIR)/$(BINARY_NAME) --mode both

# Test filtering examples
.PHONY: test-cluster
test-cluster: build
	@echo "Running cluster tests only..."
	./$(BUILD_DIR)/$(BINARY_NAME) --component cluster

.PHONY: test-mvp
test-mvp: build
	@echo "Running MVP tests only..."
	./$(BUILD_DIR)/$(BINARY_NAME) --environment mvp

.PHONY: test-gcp
test-gcp: build
	@echo "Running GCP tests only..."
	./$(BUILD_DIR)/$(BINARY_NAME) --provider gcp

.PHONY: test-happy-path
test-happy-path: build
	@echo "Running happy path tests only..."
	./$(BUILD_DIR)/$(BINARY_NAME) --category happy-path

.PHONY: test-failures
test-failures: build
	@echo "Running failure scenario tests only..."
	./$(BUILD_DIR)/$(BINARY_NAME) --category failure

.PHONY: test-specific
test-specific: build
	@echo "Running specific test e2e-001..."
	./$(BUILD_DIR)/$(BINARY_NAME) --test-id e2e-001

# Version-specific test runs
.PHONY: test-v1.0
test-v1.0: build
	@echo "Running v1.0 tests only..."
	./$(BUILD_DIR)/$(BINARY_NAME) --version v1.0

.PHONY: test-v1.1
test-v1.1: build
	@echo "Running v1.1 tests only..."
	./$(BUILD_DIR)/$(BINARY_NAME) --version v1.1

# Custom tag expressions
.PHONY: test-custom
test-custom: build
	@echo "Running tests with custom tags..."
	./$(BUILD_DIR)/$(BINARY_NAME) --tags "@cluster && @gcp && !@failure"

# Debug mode
.PHONY: test-debug
test-debug: build
	@echo "Running tests in debug mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) --debug --stop-on-failure

# Validation and utilities
.PHONY: validate-features
validate-features: build
	@echo "Validating feature files..."
	./$(BUILD_DIR)/$(BINARY_NAME) validate-features

.PHONY: list-tags
list-tags: build
	@echo "Listing available tags..."
	./$(BUILD_DIR)/$(BINARY_NAME) list-tags

.PHONY: generate-report
generate-report: build
	@echo "Generating test report..."
	./$(BUILD_DIR)/$(BINARY_NAME) generate-report

# Setup and cleanup
.PHONY: setup
setup:
	@echo "Setting up test environment..."
	@mkdir -p $(TEST_RESULTS_DIR)
	@mkdir -p $(FEATURES_DIR)
	go mod tidy

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(TEST_RESULTS_DIR)

.PHONY: clean-all
clean-all: clean
	@echo "Cleaning all generated files..."
	go clean -cache
	go clean -modcache

# Development helpers
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Testing with different outputs
.PHONY: test-junit
test-junit: build
	@mkdir -p $(TEST_RESULTS_DIR)
	./$(BUILD_DIR)/$(BINARY_NAME) --format junit > $(TEST_RESULTS_DIR)/results.xml

.PHONY: test-json
test-json: build
	@mkdir -p $(TEST_RESULTS_DIR)
	./$(BUILD_DIR)/$(BINARY_NAME) --format json > $(TEST_RESULTS_DIR)/results.json

# Docker support (if needed)
.PHONY: docker-build
docker-build:
	docker build -t hyperfleet-e2e:$(VERSION) .

.PHONY: docker-test
docker-test: docker-build
	docker run --rm \
		-e HYPERFLEET_API_URL=${HYPERFLEET_API_URL} \
		-e HYPERFLEET_AUTH_TOKEN=${HYPERFLEET_AUTH_TOKEN} \
		hyperfleet-e2e:$(VERSION)

# Environment setup examples
.PHONY: env-local
env-local:
	@echo "Setting up local environment..."
	export HYPERFLEET_API_URL=http://localhost:8080
	export FEATURE_PATH=features
	export DEBUG_MODE=true

.PHONY: env-staging
env-staging:
	@echo "Setting up staging environment..."
	export HYPERFLEET_API_URL=https://staging.hyperfleet.example.com
	export FEATURE_PATH=features

.PHONY: env-prod
env-prod:
	@echo "Setting up production environment..."
	export HYPERFLEET_API_URL=https://api.hyperfleet.example.com
	export FEATURE_PATH=features

# Help
.PHONY: help
help:
	@echo "HyperFleet E2E Test Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build           - Build the e2e-runner binary"
	@echo "  test            - Run tests with Ginkgo (default)"
	@echo "  test-godog      - Run tests with Godog"
	@echo "  test-all        - Run tests with both Ginkgo and Godog"
	@echo ""
	@echo "Filtering options:"
	@echo "  test-cluster    - Run cluster-related tests only"
	@echo "  test-mvp        - Run MVP tests only"
	@echo "  test-gcp        - Run GCP tests only"
	@echo "  test-happy-path - Run happy path tests only"
	@echo "  test-failures   - Run failure scenario tests only"
	@echo "  test-specific   - Run specific test (e2e-001)"
	@echo "  test-v1.0       - Run v1.0 tests only"
	@echo "  test-custom     - Run tests with custom tag expression"
	@echo ""
	@echo "Utilities:"
	@echo "  validate-features - Validate feature files"
	@echo "  list-tags        - List available tags"
	@echo "  generate-report  - Generate test report"
	@echo ""
	@echo "Environment variables:"
	@echo "  HYPERFLEET_API_URL      - HyperFleet API URL"
	@echo "  HYPERFLEET_AUTH_TOKEN   - Authentication token"
	@echo "  FEATURE_PATH            - Path to feature files"
	@echo "  DEBUG_MODE              - Enable debug mode (true/false)"