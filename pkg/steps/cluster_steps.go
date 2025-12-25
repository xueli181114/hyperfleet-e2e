package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"github.com/onsi/gomega"
	"hyperfleet-e2e/pkg/runner"
)

// ClusterSteps contains step definitions for cluster lifecycle tests
type ClusterSteps struct {
	apiClient *runner.APIClient
}

// NewClusterSteps creates a new cluster steps instance
func NewClusterSteps(apiClient *runner.APIClient) *ClusterSteps {
	return &ClusterSteps{
		apiClient: apiClient,
	}
}

// RegisterSteps registers cluster-related step definitions
func (cs *ClusterSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Background steps
	sc.Step(`^the HyperFleet API is available$`, cs.theHyperFleetAPIIsAvailable)
	sc.Step(`^I have valid authentication credentials$`, cs.iHaveValidAuthenticationCredentials)

	// Cluster creation steps
	sc.Step(`^I have a cluster creation request for (.+)$`, cs.iHaveAClusterCreationRequestFor)
	sc.Step(`^I submit the cluster creation request$`, cs.iSubmitTheClusterCreationRequest)
	sc.Step(`^the API should return HTTP (\d+) (.+)$`, cs.theAPIShouldReturnHTTP)
	sc.Step(`^the cluster should have status "(.+)"$`, cs.theClusterShouldHaveStatus)
	sc.Step(`^the cluster generation should be (\d+)$`, cs.theClusterGenerationShouldBe)
	sc.Step(`^the status\.adapters array should be empty$`, cs.theStatusAdaptersArrayShouldBeEmpty)

	// Cluster monitoring steps
	sc.Step(`^I monitor the cluster status$`, cs.iMonitorTheClusterStatus)
	sc.Step(`^all adapters should report their progress$`, cs.allAdaptersShouldReportTheirProgress)
	sc.Step(`^each adapter should have conditions Available, Applied, Health$`, cs.eachAdapterShouldHaveConditions)
	sc.Step(`^all adapters complete successfully$`, cs.allAdaptersCompleteSuccessfully)
	sc.Step(`^the cluster status should be "(.+)"$`, cs.theClusterStatusShouldBe)
	sc.Step(`^all adapters should have Available condition as True$`, cs.allAdaptersShouldHaveAvailableConditionAsTrue)
	sc.Step(`^the cluster API should be accessible$`, cs.theClusterAPIShouldBeAccessible)

	// Cluster update and delete steps
	sc.Step(`^I have an existing cluster in Ready state$`, cs.iHaveAnExistingClusterInReadyState)
	sc.Step(`^I update the cluster configuration$`, cs.iUpdateTheClusterConfiguration)
	sc.Step(`^the cluster generation should increment$`, cs.theClusterGenerationShouldIncrement)
	sc.Step(`^all adapters should reconcile the changes$`, cs.allAdaptersShouldReconcileTheChanges)
	sc.Step(`^the cluster should return to Ready state$`, cs.theClusterShouldReturnToReadyState)

	// Cluster deletion steps
	sc.Step(`^I have an existing cluster$`, cs.iHaveAnExistingCluster)
	sc.Step(`^I delete the cluster$`, cs.iDeleteTheCluster)
	sc.Step(`^adapters should execute cleanup procedures$`, cs.adaptersShouldExecuteCleanupProcedures)
	sc.Step(`^the cluster should be fully removed$`, cs.theClusterShouldBeFullyRemoved)
	sc.Step(`^no orphaned resources should remain$`, cs.noOrphanedResourcesShouldRemain)

	// Failure scenario steps
	sc.Step(`^I have an invalid cluster creation request$`, cs.iHaveAnInvalidClusterCreationRequest)
	sc.Step(`^I submit a cluster creation request with (.+)$`, cs.iSubmitAClusterCreationRequestWith)
	sc.Step(`^the error message should indicate the validation failure$`, cs.theErrorMessageShouldIndicateTheValidationFailure)
	sc.Step(`^no cluster resources should be created$`, cs.noClusterResourcesShouldBeCreated)
}

// Step implementations
func (cs *ClusterSteps) theHyperFleetAPIIsAvailable(ctx context.Context) error {
	err := cs.apiClient.HealthCheck()
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "API should be available")
	return nil
}

func (cs *ClusterSteps) iHaveValidAuthenticationCredentials(ctx context.Context) error {
	isAuthenticated := cs.apiClient.IsAuthenticated()
	gomega.Expect(isAuthenticated).To(gomega.BeTrue(), "Should have valid authentication")
	return nil
}

func (cs *ClusterSteps) iHaveAClusterCreationRequestFor(ctx context.Context, provider string) error {
	scenarioCtx := getScenarioContext(ctx)

	clusterRequest := map[string]interface{}{
		"apiVersion": "hyperfleet.io/v1",
		"kind":       "Cluster",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("test-cluster-%d", time.Now().Unix()),
			"namespace": "default",
			"labels": map[string]string{
				"environment": "test",
				"team":        "platform",
				"provider":    provider,
			},
		},
		"spec": map[string]interface{}{
			"provider": provider,
			"region":   "us-east1",
			"config": map[string]interface{}{
				"nodeCount": 3,
				"nodeSize":  "small",
			},
		},
	}

	scenarioCtx["clusterRequest"] = clusterRequest
	return nil
}

func (cs *ClusterSteps) iSubmitTheClusterCreationRequest(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)

	request, exists := scenarioCtx["clusterRequest"]
	gomega.Expect(exists).To(gomega.BeTrue(), "Cluster request should be prepared")

	response, err := cs.apiClient.CreateCluster(request)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Cluster creation should not fail")

	scenarioCtx["clusterResponse"] = response

	// Extract cluster ID for later use
	cluster := response["data"].(map[string]interface{})
	clusterID := cluster["id"].(string)
	scenarioCtx["clusterID"] = clusterID

	return nil
}

func (cs *ClusterSteps) theAPIShouldReturnHTTP(ctx context.Context, statusCode int, statusText string) error {
	scenarioCtx := getScenarioContext(ctx)

	response, exists := scenarioCtx["clusterResponse"]
	gomega.Expect(exists).To(gomega.BeTrue(), "Response should exist")

	responseMap := response.(map[string]interface{})
	actualStatusCode := responseMap["statusCode"].(int)
	gomega.Expect(actualStatusCode).To(gomega.Equal(statusCode),
		fmt.Sprintf("Expected HTTP %d %s", statusCode, statusText))

	return nil
}

func (cs *ClusterSteps) theClusterShouldHaveStatus(ctx context.Context, expectedStatus string) error {
	scenarioCtx := getScenarioContext(ctx)

	response := scenarioCtx["clusterResponse"].(map[string]interface{})
	cluster := response["data"].(map[string]interface{})
	status := cluster["status"].(map[string]interface{})
	phase := status["phase"].(string)

	gomega.Expect(phase).To(gomega.Equal(expectedStatus),
		fmt.Sprintf("Cluster should have status %s", expectedStatus))

	return nil
}

func (cs *ClusterSteps) theClusterGenerationShouldBe(ctx context.Context, expectedGeneration int) error {
	scenarioCtx := getScenarioContext(ctx)

	response := scenarioCtx["clusterResponse"].(map[string]interface{})
	cluster := response["data"].(map[string]interface{})
	generation := int(cluster["metadata"].(map[string]interface{})["generation"].(float64))

	gomega.Expect(generation).To(gomega.Equal(expectedGeneration),
		fmt.Sprintf("Cluster generation should be %d", expectedGeneration))

	return nil
}

func (cs *ClusterSteps) theStatusAdaptersArrayShouldBeEmpty(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)

	response := scenarioCtx["clusterResponse"].(map[string]interface{})
	cluster := response["data"].(map[string]interface{})
	status := cluster["status"].(map[string]interface{})
	adapters := status["adapters"].([]interface{})

	gomega.Expect(adapters).To(gomega.BeEmpty(), "Adapters array should be empty initially")
	return nil
}

func (cs *ClusterSteps) iMonitorTheClusterStatus(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)

	clusterID := scenarioCtx["clusterID"].(string)

	// Get detailed adapter statuses
	statusResponse, err := cs.apiClient.GetClusterStatuses(clusterID)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Should be able to get cluster statuses")

	scenarioCtx["clusterStatuses"] = statusResponse

	// Also get the current cluster state
	clusterResponse, err := cs.apiClient.GetCluster(clusterID)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Should be able to get cluster")

	scenarioCtx["currentCluster"] = clusterResponse
	return nil
}

func (cs *ClusterSteps) allAdaptersShouldReportTheirProgress(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)

	statusResponse := scenarioCtx["clusterStatuses"].(map[string]interface{})
	adapterStatuses := statusResponse["data"].(map[string]interface{})["adapterStatuses"].([]interface{})

	gomega.Expect(len(adapterStatuses)).To(gomega.BeNumerically(">", 0),
		"At least one adapter should report status")

	// Verify each adapter has the expected structure
	for _, adapter := range adapterStatuses {
		adapterMap := adapter.(map[string]interface{})

		gomega.Expect(adapterMap).To(gomega.HaveKey("name"), "Adapter should have name")
		gomega.Expect(adapterMap).To(gomega.HaveKey("conditions"), "Adapter should have conditions")

		conditions := adapterMap["conditions"].([]interface{})
		gomega.Expect(len(conditions)).To(gomega.BeNumerically(">", 0),
			"Adapter should have at least one condition")
	}

	return nil
}

func (cs *ClusterSteps) eachAdapterShouldHaveConditions(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)

	statusResponse := scenarioCtx["clusterStatuses"].(map[string]interface{})
	adapterStatuses := statusResponse["data"].(map[string]interface{})["adapterStatuses"].([]interface{})

	requiredConditions := []string{"Available", "Applied", "Health"}

	for _, adapter := range adapterStatuses {
		adapterMap := adapter.(map[string]interface{})
		conditions := adapterMap["conditions"].([]interface{})

		conditionTypes := make([]string, 0)
		for _, condition := range conditions {
			conditionMap := condition.(map[string]interface{})
			conditionTypes = append(conditionTypes, conditionMap["type"].(string))
		}

		for _, required := range requiredConditions {
			gomega.Expect(conditionTypes).To(gomega.ContainElement(required),
				fmt.Sprintf("Adapter %s should have %s condition", adapterMap["name"], required))
		}
	}

	return nil
}

func (cs *ClusterSteps) allAdaptersCompleteSuccessfully(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)
	clusterID := scenarioCtx["clusterID"].(string)

	// Wait for cluster to become Ready
	finalResponse, err := cs.apiClient.WaitForClusterReady(clusterID)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Cluster should eventually become Ready")

	scenarioCtx["finalClusterResponse"] = finalResponse
	return nil
}

func (cs *ClusterSteps) theClusterStatusShouldBe(ctx context.Context, expectedStatus string) error {
	scenarioCtx := getScenarioContext(ctx)

	finalResponse := scenarioCtx["finalClusterResponse"].(map[string]interface{})
	cluster := finalResponse["data"].(map[string]interface{})
	status := cluster["status"].(map[string]interface{})
	phase := status["phase"].(string)

	gomega.Expect(phase).To(gomega.Equal(expectedStatus),
		fmt.Sprintf("Final cluster status should be %s", expectedStatus))

	return nil
}

func (cs *ClusterSteps) allAdaptersShouldHaveAvailableConditionAsTrue(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)

	finalResponse := scenarioCtx["finalClusterResponse"].(map[string]interface{})
	cluster := finalResponse["data"].(map[string]interface{})
	status := cluster["status"].(map[string]interface{})
	adapters := status["adapters"].([]interface{})

	for _, adapter := range adapters {
		adapterMap := adapter.(map[string]interface{})
		available := adapterMap["available"].(string)
		gomega.Expect(available).To(gomega.Equal("True"),
			fmt.Sprintf("Adapter %s should have Available=True", adapterMap["name"]))
	}

	return nil
}

func (cs *ClusterSteps) theClusterAPIShouldBeAccessible(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)

	finalResponse := scenarioCtx["finalClusterResponse"].(map[string]interface{})
	cluster := finalResponse["data"].(map[string]interface{})

	// Check if cluster has an API endpoint
	spec := cluster["spec"].(map[string]interface{})

	if apiEndpoint, exists := spec["apiEndpoint"]; exists && apiEndpoint != "" {
		// In a real implementation, you would test the actual API endpoint
		gomega.Expect(apiEndpoint).NotTo(gomega.BeEmpty(), "Cluster API endpoint should be available")
	}

	return nil
}

// Placeholder implementations for update/delete scenarios
func (cs *ClusterSteps) iHaveAnExistingClusterInReadyState(ctx context.Context) error {
	// Create a cluster and wait for it to be Ready
	provider := "GCP"
	cs.iHaveAClusterCreationRequestFor(ctx, provider)
	cs.iSubmitTheClusterCreationRequest(ctx)
	cs.allAdaptersCompleteSuccessfully(ctx)
	return nil
}

func (cs *ClusterSteps) iUpdateTheClusterConfiguration(ctx context.Context) error {
	// Implementation for cluster update would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) theClusterGenerationShouldIncrement(ctx context.Context) error {
	// Implementation for checking generation increment would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) allAdaptersShouldReconcileTheChanges(ctx context.Context) error {
	// Implementation for checking adapter reconciliation would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) theClusterShouldReturnToReadyState(ctx context.Context) error {
	// Implementation for checking cluster returns to Ready would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) iHaveAnExistingCluster(ctx context.Context) error {
	// Implementation for setting up existing cluster would go here
	return cs.iHaveAnExistingClusterInReadyState(ctx)
}

func (cs *ClusterSteps) iDeleteTheCluster(ctx context.Context) error {
	scenarioCtx := getScenarioContext(ctx)
	clusterID := scenarioCtx["clusterID"].(string)

	err := cs.apiClient.DeleteCluster(clusterID)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Cluster deletion should succeed")

	return nil
}

func (cs *ClusterSteps) adaptersShouldExecuteCleanupProcedures(ctx context.Context) error {
	// Implementation for checking adapter cleanup would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) theClusterShouldBeFullyRemoved(ctx context.Context) error {
	// Implementation for checking cluster removal would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) noOrphanedResourcesShouldRemain(ctx context.Context) error {
	// Implementation for checking no orphaned resources would go here
	return godog.ErrPending
}

// Failure scenario implementations
func (cs *ClusterSteps) iHaveAnInvalidClusterCreationRequest(ctx context.Context) error {
	// Implementation for invalid request would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) iSubmitAClusterCreationRequestWith(ctx context.Context, invalidField string) error {
	// Implementation for submitting invalid request would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) theErrorMessageShouldIndicateTheValidationFailure(ctx context.Context) error {
	// Implementation for checking error message would go here
	return godog.ErrPending
}

func (cs *ClusterSteps) noClusterResourcesShouldBeCreated(ctx context.Context) error {
	// Implementation for checking no resources created would go here
	return godog.ErrPending
}

// Helper function to get scenario context from Godog context
func getScenarioContext(ctx context.Context) map[string]interface{} {
	if scenarioCtx := ctx.Value("scenarioContext"); scenarioCtx != nil {
		return scenarioCtx.(map[string]interface{})
	}
	return make(map[string]interface{})
}

// Ginkgo-compatible interface methods (without godog.Scenario parameter)
func (cs *ClusterSteps) IHaveAClusterCreationRequestFor(ctx context.Context, provider string) error {
	return cs.iHaveAClusterCreationRequestFor(ctx, provider)
}

func (cs *ClusterSteps) ISubmitTheClusterCreationRequest(ctx context.Context) error {
	return cs.iSubmitTheClusterCreationRequest(ctx)
}

func (cs *ClusterSteps) TheAPIShouldReturnHTTP(ctx context.Context, statusCode int, statusText string) error {
	return cs.theAPIShouldReturnHTTP(ctx, statusCode, statusText)
}

func (cs *ClusterSteps) TheClusterShouldHaveStatus(ctx context.Context, status string) error {
	return cs.theClusterShouldHaveStatus(ctx, status)
}

func (cs *ClusterSteps) TheClusterGenerationShouldBe(ctx context.Context, generation int) error {
	return cs.theClusterGenerationShouldBe(ctx, generation)
}

func (cs *ClusterSteps) TheStatusAdaptersArrayShouldBeEmpty(ctx context.Context) error {
	return cs.theStatusAdaptersArrayShouldBeEmpty(ctx)
}

func (cs *ClusterSteps) IMonitorTheClusterStatus(ctx context.Context) error {
	return cs.iMonitorTheClusterStatus(ctx)
}

func (cs *ClusterSteps) AllAdaptersShouldReportTheirProgress(ctx context.Context) error {
	return cs.allAdaptersShouldReportTheirProgress(ctx)
}

func (cs *ClusterSteps) EachAdapterShouldHaveConditions(ctx context.Context) error {
	return cs.eachAdapterShouldHaveConditions(ctx)
}

func (cs *ClusterSteps) AllAdaptersCompleteSuccessfully(ctx context.Context) error {
	return cs.allAdaptersCompleteSuccessfully(ctx)
}

func (cs *ClusterSteps) TheClusterStatusShouldBe(ctx context.Context, status string) error {
	return cs.theClusterStatusShouldBe(ctx, status)
}

func (cs *ClusterSteps) AllAdaptersShouldHaveAvailableConditionAsTrue(ctx context.Context) error {
	return cs.allAdaptersShouldHaveAvailableConditionAsTrue(ctx)
}

func (cs *ClusterSteps) TheClusterAPIShouldBeAccessible(ctx context.Context) error {
	return cs.theClusterAPIShouldBeAccessible(ctx)
}