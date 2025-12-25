# HyperFleet E2E Testing Framework

A comprehensive Gherkin-based E2E testing framework that integrates **Cucumber/Godog** with **Ginkgo/Gomega**, featuring automatic tag mapping and version control for test scenarios.

## Features

- ✅ **Gherkin Feature Files**: Write tests in natural language using Cucumber syntax
- ✅ **Automatic Tag Mapping**: Ginkgo labels are automatically derived from Gherkin tags
- ✅ **Version Control**: Tag tests by version (`@v1.0`, `@v1.1`) for progressive testing
- ✅ **Flexible Filtering**: Run specific subsets of tests based on components, actions, providers, etc.
- ✅ **Dual Runner Support**: Run tests with both Ginkgo and Godog
- ✅ **Rich CLI**: Command-line interface with extensive filtering options

## Project Structure

```
hyperfleet-e2e/
├── features/                     # Gherkin feature files
│   ├── cluster_lifecycle.feature
│   └── nodepool_lifecycle.feature
├── pkg/
│   ├── tags/                     # Tag parsing and mapping
│   │   └── version_tags.go
│   ├── runner/                   # Test runners and integration
│   │   ├── gherkin_runner.go
│   │   ├── step_registry.go
│   │   └── api_client.go
│   └── steps/                    # Step definitions
│       └── cluster_steps.go
├── tests/                        # Test entry points
│   └── e2e_test.go
├── cmd/
│   └── e2e-runner/              # CLI application
│       └── main.go
├── config.yaml                  # Configuration file
├── Makefile                     # Build and test automation
└── README.md                    # This file
```

## Quick Start

### 1. Install Dependencies

```bash
make deps
```

### 2. Build the Test Runner

```bash
make build
```

### 3. Run Tests

```bash
# Run all tests with Ginkgo
make test

# Run specific test categories
make test-cluster      # Only cluster tests
make test-mvp         # Only MVP tests
make test-gcp         # Only GCP tests
make test-failures    # Only failure scenarios
```

## Tag System

### Automatic Tag Mapping

Gherkin tags are automatically mapped to Ginkgo labels, enabling powerful test filtering:

| Gherkin Tag | Ginkgo Label | Description |
|-------------|--------------|-------------|
| `@v1.0` | `v1.0, v1` | Version 1.0 tests + major version |
| `@cluster` | `cluster` | Cluster-related tests |
| `@create` | `create, cluster-create` | Create action + component-action |
| `@gcp` | `gcp, provider-gcp` | GCP tests + provider prefix |
| `@happy-path` | `happy-path` | Successful scenarios |
| `@e2e-001` | `e2e-001` | Specific test identifier |

### Tag Categories

#### Version Tags
- `@v1.0`, `@v1.1`, `@v2.0` - Version control for progressive testing

#### Component Tags
- `@cluster` - Cluster lifecycle tests
- `@nodepool` - NodePool management tests
- `@adapter` - Adapter-specific tests
- `@api` - API-level tests
- `@sentinel` - Sentinel operator tests

#### Action Tags
- `@create` - Resource creation tests
- `@update` - Resource modification tests
- `@delete` - Resource deletion tests
- `@validation` - Input validation tests

#### Category Tags
- `@happy-path` - Successful execution scenarios
- `@failure` - Error and failure scenarios
- `@business-logic-failure` - Business logic validation failures
- `@infrastructure-failure` - Infrastructure/platform failures
- `@lifecycle` - Complete lifecycle tests

#### Provider Tags
- `@gcp` - Google Cloud Platform tests
- `@aws` - Amazon Web Services tests
- `@azure` - Microsoft Azure tests

#### Environment Tags
- `@mvp` - MVP (Minimum Viable Product) scope
- `@post-mvp` - Post-MVP features

#### Test ID Tags
- `@e2e-001` - Happy path cluster creation
- `@e2e-fail-001` - Validation failures
- `@e2e-fail-002` - Infrastructure failures

## Example Feature File

```gherkin
@v1.0 @cluster @lifecycle @mvp
Feature: HyperFleet Cluster Lifecycle Management
  As a platform engineer
  I want to create and manage clusters through the HyperFleet API
  So that I can provision infrastructure for my applications

  @v1.0 @cluster @create @gcp @happy-path @e2e-001
  Scenario: Full Cluster Creation Flow on GCP
    Given the HyperFleet API is available
    And I have valid authentication credentials
    And I have a cluster creation request for GCP
      | field       | value                           |
      | provider    | GCP                            |
      | region      | us-east1                       |
      | environment | test                           |
      | team        | platform                       |
    When I submit the cluster creation request
    Then the API should return HTTP 201 Created
    And the cluster should have status "Not Ready"
    And the cluster generation should be 1
    When I monitor the cluster status
    Then all adapters should report their progress
    And each adapter should have conditions Available, Applied, Health
    When all adapters complete successfully
    Then the cluster status should be "Ready"
    And all adapters should have Available condition as True
    And the cluster API should be accessible
```

## CLI Usage

### Basic Commands

```bash
# Build and run tests
./build/e2e-runner

# Run with specific mode
./build/e2e-runner --mode ginkgo    # Ginkgo only
./build/e2e-runner --mode godog     # Godog only
./build/e2e-runner --mode both      # Both runners
```

### Filtering Options

```bash
# Filter by version
./build/e2e-runner --version v1.0

# Filter by component
./build/e2e-runner --component cluster

# Filter by action
./build/e2e-runner --action create

# Filter by category
./build/e2e-runner --category happy-path

# Filter by provider
./build/e2e-runner --provider gcp

# Filter by environment
./build/e2e-runner --environment mvp

# Filter by test ID
./build/e2e-runner --test-id e2e-001

# Custom tag expression
./build/e2e-runner --tags "@cluster && @gcp && !@failure"
```

### Configuration

```bash
# Set API endpoint
./build/e2e-runner --api-url https://api.hyperfleet.example.com

# Enable debug mode
./build/e2e-runner --debug --stop-on-failure

# Change output format
./build/e2e-runner --format junit
./build/e2e-runner --format json
```

### Utility Commands

```bash
# List all available tags
./build/e2e-runner list-tags

# Validate feature files
./build/e2e-runner validate-features

# Generate test report
./build/e2e-runner generate-report
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HYPERFLEET_API_URL` | HyperFleet API endpoint | `http://localhost:8080` |
| `HYPERFLEET_AUTH_TOKEN` | Authentication token | (required) |
| `FEATURE_PATH` | Path to feature files | `features` |
| `DEBUG_MODE` | Enable debug output | `false` |
| `TEST_TIMEOUT` | Test timeout duration | `30m` |
| `TEST_CONCURRENCY` | Concurrent test runners | `1` |

## Make Targets

### Basic Operations
```bash
make build          # Build the test runner
make test           # Run tests with Ginkgo
make test-godog     # Run tests with Godog
make test-all       # Run with both runners
```

### Filtered Test Runs
```bash
make test-cluster      # Cluster tests only
make test-mvp         # MVP features only
make test-gcp         # GCP provider only
make test-happy-path  # Successful scenarios only
make test-failures    # Failure scenarios only
make test-v1.0        # Version 1.0 tests only
```

### Development
```bash
make deps           # Install dependencies
make fmt            # Format code
make lint           # Run linter
make vet            # Run go vet
make clean          # Clean build artifacts
```

This framework provides a robust foundation for E2E testing with excellent tag management and automatic Ginkgo integration!
