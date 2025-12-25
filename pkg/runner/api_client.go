package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/onsi/gomega"
)

// APIClient handles HTTP communication with the HyperFleet API
type APIClient struct {
	BaseURL    string
	AuthToken  string
	HTTPClient *http.Client
}

// NewAPIClient creates a new API client instance
func NewAPIClient(baseURL, authToken string) *APIClient {
	return &APIClient{
		BaseURL:   baseURL,
		AuthToken: authToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HealthCheck verifies API connectivity
func (c *APIClient) HealthCheck() error {
	resp, err := c.doRequest("GET", "/health", nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// IsAuthenticated checks if the client has valid authentication
func (c *APIClient) IsAuthenticated() bool {
	return c.AuthToken != ""
}

// CreateCluster creates a new cluster via the API
func (c *APIClient) CreateCluster(request interface{}) (map[string]interface{}, error) {
	resp, err := c.doRequest("POST", "/api/hyperfleet/v1/clusters", request)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Add status code to result for validation
	result["statusCode"] = resp.StatusCode

	return result, nil
}

// GetCluster retrieves cluster details by ID
func (c *APIClient) GetCluster(clusterID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/hyperfleet/v1/clusters/%s", clusterID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result["statusCode"] = resp.StatusCode
	return result, nil
}

// GetClusterStatuses retrieves detailed adapter statuses for a cluster
func (c *APIClient) GetClusterStatuses(clusterID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/hyperfleet/v1/clusters/%s/statuses", clusterID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster statuses: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result["statusCode"] = resp.StatusCode
	return result, nil
}

// ListClusters retrieves all clusters with optional filtering
func (c *APIClient) ListClusters(filters map[string]string) (map[string]interface{}, error) {
	path := "/api/hyperfleet/v1/clusters"

	// Add query parameters for filtering if provided
	if len(filters) > 0 {
		path += "?"
		for key, value := range filters {
			path += fmt.Sprintf("%s=%s&", key, value)
		}
		path = path[:len(path)-1] // Remove trailing &
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result["statusCode"] = resp.StatusCode
	return result, nil
}

// CreateNodePool creates a new nodepool for a cluster
func (c *APIClient) CreateNodePool(clusterID string, request interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/hyperfleet/v1/clusters/%s/nodepools", clusterID)
	resp, err := c.doRequest("POST", path, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create nodepool: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result["statusCode"] = resp.StatusCode
	return result, nil
}

// GetNodePool retrieves nodepool details by ID
func (c *APIClient) GetNodePool(clusterID, nodepoolID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/hyperfleet/v1/clusters/%s/nodepools/%s", clusterID, nodepoolID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodepool: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result["statusCode"] = resp.StatusCode
	return result, nil
}

// ListNodePools retrieves all nodepools for a cluster
func (c *APIClient) ListNodePools(clusterID string, filters map[string]string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/hyperfleet/v1/clusters/%s/nodepools", clusterID)

	// Add query parameters for filtering if provided
	if len(filters) > 0 {
		path += "?"
		for key, value := range filters {
			path += fmt.Sprintf("%s=%s&", key, value)
		}
		path = path[:len(path)-1] // Remove trailing &
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodepools: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result["statusCode"] = resp.StatusCode
	return result, nil
}

// WaitForClusterReady polls cluster status until it becomes Ready or times out
func (c *APIClient) WaitForClusterReady(clusterID string) (map[string]interface{}, error) {
	timeout := 10 * time.Minute
	pollInterval := 15 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		cluster, err := c.GetCluster(clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster status: %w", err)
		}

		clusterData := cluster["cluster"].(map[string]interface{})
		status := clusterData["status"].(map[string]interface{})
		phase := status["phase"].(string)

		if phase == "Ready" {
			return cluster, nil
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("timeout waiting for cluster %s to become Ready", clusterID)
}

// WaitForNodePoolReady polls nodepool status until it becomes Ready or times out
func (c *APIClient) WaitForNodePoolReady(clusterID, nodepoolID string) (map[string]interface{}, error) {
	timeout := 10 * time.Minute
	pollInterval := 15 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		nodepool, err := c.GetNodePool(clusterID, nodepoolID)
		if err != nil {
			return nil, fmt.Errorf("failed to get nodepool status: %w", err)
		}

		nodepoolData := nodepool["nodepool"].(map[string]interface{})
		status := nodepoolData["status"].(map[string]interface{})
		phase := status["phase"].(string)

		if phase == "Ready" {
			return nodepool, nil
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("timeout waiting for nodepool %s to become Ready", nodepoolID)
}

// doRequest performs an HTTP request with proper authentication
func (c *APIClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header if token is available
	if c.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))
	}

	// Set content type for POST/PUT requests
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add user agent
	req.Header.Set("User-Agent", "hyperfleet-e2e-tests/1.0")

	return c.HTTPClient.Do(req)
}

// DeleteCluster deletes a cluster by ID
func (c *APIClient) DeleteCluster(clusterID string) error {
	path := fmt.Sprintf("/api/hyperfleet/v1/clusters/%s", clusterID)
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete cluster failed with status: %d", resp.StatusCode)
	}

	return nil
}

// DeleteNodePool deletes a nodepool by ID
func (c *APIClient) DeleteNodePool(clusterID, nodepoolID string) error {
	path := fmt.Sprintf("/api/hyperfleet/v1/clusters/%s/nodepools/%s", clusterID, nodepoolID)
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete nodepool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete nodepool failed with status: %d", resp.StatusCode)
	}

	return nil
}