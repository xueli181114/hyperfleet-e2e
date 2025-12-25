@v1.0 @cluster @lifecycle @mvp
Feature: HyperFleet Cluster Lifecycle Management
  As a platform engineer
  I want to create and manage clusters through the HyperFleet API
  So that I can provision infrastructure for my applications

  Background:
    Given the HyperFleet API is available
    And I have valid authentication credentials

  @v1.0 @cluster @create @gcp @happy-path @e2e-001
  Scenario: Full Cluster Creation Flow on GCP
    Given I have a cluster creation request for GCP
      | field       | value                           |
      | provider    | GCP                            |
      | region      | us-east1                       |
      | environment | test                           |
      | team        | platform                       |
    When I submit the cluster creation request
    Then the API should return HTTP 201 Created
    And the cluster should have status "Not Ready"
    And the cluster generation should be 1
    And the status.adapters array should be empty
    When I monitor the cluster status
    Then all adapters should report their progress
    And each adapter should have conditions Available, Applied, Health
    When all adapters complete successfully
    Then the cluster status should be "Ready"
    And all adapters should have Available condition as True
    And the cluster API should be accessible

  @v1.1 @cluster @update @post-mvp @e2e-002
  Scenario: Cluster Configuration Update
    Given I have an existing cluster in Ready state
    When I update the cluster configuration
    Then the cluster generation should increment
    And all adapters should reconcile the changes
    And the cluster should return to Ready state

  @v1.1 @cluster @delete @post-mvp @e2e-003
  Scenario: Cluster Deletion
    Given I have an existing cluster
    When I delete the cluster
    Then adapters should execute cleanup procedures
    And the cluster should be fully removed
    And no orphaned resources should remain

  @v1.0 @cluster @validation @failure @e2e-fail-001
  Scenario Outline: Cluster API Request Validation Failures
    Given I have an invalid cluster creation request
    When I submit a cluster creation request with <invalid_field>
    Then the API should return HTTP 400 Bad Request
    And the error message should indicate the validation failure
    And no cluster resources should be created

    Examples:
      | invalid_field           |
      | missing name            |
      | unsupported field       |
      | invalid region type     |
      | existing cluster name   |
      | invalid JSON syntax     |

  @v1.0 @adapter @business-logic-failure @failure @e2e-fail-003
  Scenario: Adapter Business Logic Failure
    Given I have a cluster creation request with missing prerequisites
    When I submit the cluster creation request
    Then the validation adapter should report failure
    And the Health condition should be True
    And the Available condition should be False
    And the failure reason should be clearly indicated

  @v1.0 @adapter @infrastructure-failure @failure @e2e-fail-004
  Scenario: Adapter Infrastructure Failure - Job Creation Failure
    Given an adapter with invalid YAML configuration
    When the adapter attempts to create a Kubernetes Job
    Then the Job creation should fail
    And the Applied condition should be False
    And the Health condition should be False
    And detailed error information should be provided

  @v1.0 @adapter @execution-failure @failure @e2e-fail-004
  Scenario: Adapter Infrastructure Failure - Job Execution Failure
    Given an adapter with incorrect runtime parameters
    When the Kubernetes Job executes
    Then the Job should fail with non-zero exit code
    And the Applied condition should be True
    And the Available condition should be False
    And the Health condition should be False