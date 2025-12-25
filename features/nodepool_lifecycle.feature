@v1.0 @nodepool @lifecycle @mvp
Feature: HyperFleet NodePool Lifecycle Management
  As a platform engineer
  I want to create and manage nodepools within clusters
  So that I can provision compute resources for workloads

  Background:
    Given the HyperFleet API is available
    And I have valid authentication credentials
    And I have an existing cluster in Ready state

  @v1.0 @nodepool @create @happy-path @e2e-004
  Scenario: Full NodePool Creation Flow
    Given I have a nodepool creation request
      | field       | value         |
      | name        | gpu-nodepool  |
      | machineType | n1-standard-8 |
      | replicas    | 2             |
      | workload    | gpu           |
      | tier        | compute       |
    When I submit the nodepool creation request
    Then the API should return HTTP 201 Created
    And the nodepool should have status "Not Ready"
    And the nodepool generation should be 1
    When I check the nodepool list
    Then the nodepool should appear in the response
    And I should be able to filter by labels
    When I monitor the nodepool status
    Then all adapters should report their progress
    And the validation adapter should complete successfully
    And the nodepool adapter should complete successfully
    When all adapters complete successfully
    Then the nodepool status should be "Ready"
    And the nodes should be running and joined to the cluster

  @v1.1 @nodepool @update @post-mvp @e2e-005
  Scenario: NodePool Configuration Update
    Given I have an existing nodepool in Ready state
    When I update the nodepool replica count
    Then the nodepool generation should increment
    And all adapters should reconcile the changes
    And the nodepool should return to Ready state
    And the correct number of nodes should be running

  @v1.1 @nodepool @delete @post-mvp @e2e-006
  Scenario: NodePool Deletion
    Given I have an existing nodepool
    When I delete the nodepool
    Then adapters should execute cleanup procedures
    And the nodepool should be fully removed
    And no orphaned nodes should remain

  @v1.0 @nodepool @validation @failure @e2e-fail-002
  Scenario Outline: NodePool API Request Validation Failures
    Given I have an invalid nodepool creation request
    When I submit a nodepool creation request with <invalid_field>
    Then the API should return HTTP <status_code>
    And the error message should indicate the validation failure
    And no nodepool resources should be created

    Examples:
      | invalid_field           | status_code |
      | missing name            | 400         |
      | negative replicas       | 400         |
      | empty machineType       | 400         |
      | invalid JSON syntax     | 400         |
      | non-existent cluster    | 404         |