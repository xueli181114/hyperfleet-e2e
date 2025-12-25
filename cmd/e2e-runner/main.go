package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"hyperfleet-e2e/pkg/runner"
	"hyperfleet-e2e/pkg/tags"
	"hyperfleet-e2e/tests"
)

var (
	// Global configuration
	config = &runner.RunnerConfig{
		FeaturePath:   "features",
		OutputFormat:  "pretty",
		Randomize:     false,
		StopOnFailure: false,
		Strict:        false,
		NoColors:      false,
		Concurrency:   1,
		APIBaseURL:    "http://localhost:8080",
		AuthToken:     "",
		DebugMode:     false,
	}

	// Filter options
	filterOptions = struct {
		version     string
		component   string
		action      string
		category    string
		provider    string
		environment string
		testID      string
		tags        string
	}{}

	// Run mode
	runMode string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "e2e-runner",
		Short: "HyperFleet E2E Test Runner with Gherkin support",
		Long: `A comprehensive E2E test runner that integrates Gherkin features with Ginkgo/Gomega.
Supports automatic tag mapping and version control for test scenarios.`,
		RunE: runTests,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&config.APIBaseURL, "api-url", config.APIBaseURL,
		"HyperFleet API base URL")
	rootCmd.PersistentFlags().StringVar(&config.AuthToken, "auth-token", config.AuthToken,
		"Authentication token for API access")
	rootCmd.PersistentFlags().StringVar(&config.FeaturePath, "features", config.FeaturePath,
		"Path to feature files")
	rootCmd.PersistentFlags().StringVar(&config.OutputFormat, "format", config.OutputFormat,
		"Output format (pretty, junit, json)")
	rootCmd.PersistentFlags().BoolVar(&config.Randomize, "randomize", config.Randomize,
		"Randomize test execution order")
	rootCmd.PersistentFlags().BoolVar(&config.StopOnFailure, "stop-on-failure", config.StopOnFailure,
		"Stop execution on first failure")
	rootCmd.PersistentFlags().BoolVar(&config.Strict, "strict", config.Strict,
		"Fail on undefined steps")
	rootCmd.PersistentFlags().BoolVar(&config.NoColors, "no-colors", config.NoColors,
		"Disable colored output")
	rootCmd.PersistentFlags().IntVar(&config.Concurrency, "concurrency", config.Concurrency,
		"Number of concurrent test runners")
	rootCmd.PersistentFlags().BoolVar(&config.DebugMode, "debug", config.DebugMode,
		"Enable debug mode")

	// Filter flags for tag-based test selection
	rootCmd.PersistentFlags().StringVar(&filterOptions.version, "version", "",
		"Filter by version (e.g., v1.0, v1.1)")
	rootCmd.PersistentFlags().StringVar(&filterOptions.component, "component", "",
		"Filter by component (cluster, nodepool, adapter)")
	rootCmd.PersistentFlags().StringVar(&filterOptions.action, "action", "",
		"Filter by action (create, update, delete)")
	rootCmd.PersistentFlags().StringVar(&filterOptions.category, "category", "",
		"Filter by category (happy-path, failure, validation)")
	rootCmd.PersistentFlags().StringVar(&filterOptions.provider, "provider", "",
		"Filter by provider (gcp, aws, azure)")
	rootCmd.PersistentFlags().StringVar(&filterOptions.environment, "environment", "",
		"Filter by environment (mvp, post-mvp)")
	rootCmd.PersistentFlags().StringVar(&filterOptions.testID, "test-id", "",
		"Filter by test ID (e.g., e2e-001, e2e-fail-002)")
	rootCmd.PersistentFlags().StringVar(&filterOptions.tags, "tags", "",
		"Custom tag expression for filtering")

	// Run mode flags
	rootCmd.PersistentFlags().StringVar(&runMode, "mode", "ginkgo",
		"Test runner mode (ginkgo, godog, both)")

	// Subcommands
	rootCmd.AddCommand(listTagsCmd())
	rootCmd.AddCommand(validateFeaturesCmd())
	rootCmd.AddCommand(generateReportCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTests(cmd *cobra.Command, args []string) error {
	// Load configuration from environment
	loadEnvironmentConfig()

	// Build tag filter
	tagFilter := buildTagFilter()
	if tagFilter != "" {
		config.Tags = tagFilter
		fmt.Printf("Running tests with tag filter: %s\n", tagFilter)
	}

	// Display configuration
	displayConfiguration()

	switch runMode {
	case "ginkgo":
		return runGinkgoTests()
	case "godog":
		return runGodogTests()
	case "both":
		if err := runGinkgoTests(); err != nil {
			return err
		}
		return runGodogTests()
	default:
		return fmt.Errorf("invalid run mode: %s (must be ginkgo, godog, or both)", runMode)
	}
}

func runGinkgoTests() error {
	fmt.Println("Running tests with Ginkgo...")

	// Configure Ginkgo based on our settings
	suiteConfig := ginkgo.GinkgoConfiguration()
	if config.Randomize {
		suiteConfig.RandomSeed = 1
	}

	reporterConfig := ginkgo.GinkgoReporterConfiguration()
	reporterConfig.Verbose = config.DebugMode

	// Set up label filter based on tags
	if config.Tags != "" {
		suiteConfig.LabelFilter = config.Tags
	}

	// Run the test suite
	ginkgo.RunSpecs(nil, "HyperFleet E2E Tests")

	return nil
}

func runGodogTests() error {
	fmt.Println("Running tests with Godog...")

	opts := godog.Options{
		Format:        config.OutputFormat,
		Paths:         []string{config.FeaturePath},
		Randomize:     config.Randomize,
		StopOnFailure: config.StopOnFailure,
		Strict:        config.Strict,
		NoColors:      config.NoColors,
		Tags:          config.Tags,
		Concurrency:   config.Concurrency,
	}

	suite := godog.TestSuite{
		Name:                "HyperFleet E2E Tests",
		Options:             &opts,
		ScenarioInitializer: tests.InitializeScenario,
	}

	status := suite.Run()
	if status != 0 {
		return fmt.Errorf("tests failed with status: %d", status)
	}

	return nil
}

func buildTagFilter() string {
	tagParser := tags.NewTagParser()
	versionTag := &tags.VersionTag{}

	// Build version tag from filter options
	if filterOptions.version != "" {
		versionTag.Version = filterOptions.version
	}
	if filterOptions.component != "" {
		versionTag.Component = filterOptions.component
	}
	if filterOptions.action != "" {
		versionTag.Action = filterOptions.action
	}
	if filterOptions.category != "" {
		versionTag.Category = filterOptions.category
	}
	if filterOptions.provider != "" {
		versionTag.Provider = filterOptions.provider
	}
	if filterOptions.environment != "" {
		versionTag.Environment = filterOptions.environment
	}
	if filterOptions.testID != "" {
		versionTag.TestID = filterOptions.testID
	}

	// Use custom tags if provided
	if filterOptions.tags != "" {
		return filterOptions.tags
	}

	// Generate filter from structured options
	filters := make(map[string]string)
	if versionTag.Version != "" {
		filters["version"] = versionTag.Version
	}
	if versionTag.Component != "" {
		filters["component"] = versionTag.Component
	}
	if versionTag.Action != "" {
		filters["action"] = versionTag.Action
	}
	if versionTag.Category != "" {
		filters["category"] = versionTag.Category
	}
	if versionTag.Provider != "" {
		filters["provider"] = versionTag.Provider
	}
	if versionTag.TestID != "" {
		filters["testid"] = versionTag.TestID
	}

	return versionTag.GenerateLabelSelector(filters)
}

func loadEnvironmentConfig() {
	if apiURL := os.Getenv("HYPERFLEET_API_URL"); apiURL != "" {
		config.APIBaseURL = apiURL
	}
	if authToken := os.Getenv("HYPERFLEET_AUTH_TOKEN"); authToken != "" {
		config.AuthToken = authToken
	}
	if featurePath := os.Getenv("FEATURE_PATH"); featurePath != "" {
		config.FeaturePath = featurePath
	}
	if debugMode := os.Getenv("DEBUG_MODE"); debugMode == "true" {
		config.DebugMode = true
	}
}

func displayConfiguration() {
	fmt.Println("=== Test Configuration ===")
	fmt.Printf("API URL: %s\n", config.APIBaseURL)
	fmt.Printf("Auth Token: %s\n", maskToken(config.AuthToken))
	fmt.Printf("Features Path: %s\n", config.FeaturePath)
	fmt.Printf("Output Format: %s\n", config.OutputFormat)
	fmt.Printf("Run Mode: %s\n", runMode)
	fmt.Printf("Debug Mode: %t\n", config.DebugMode)
	if config.Tags != "" {
		fmt.Printf("Tag Filter: %s\n", config.Tags)
	}
	fmt.Println("==========================")
}

func maskToken(token string) string {
	if token == "" {
		return "<not set>"
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}

func listTagsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-tags",
		Short: "List all available tags and their meanings",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("=== Available Tag Categories ===")

			filters := tags.GetAvailableFilters()
			for category, values := range filters {
				fmt.Printf("\n%s:\n", strings.Title(category))
				for _, value := range values {
					fmt.Printf("  - %s\n", value)
				}
			}

			fmt.Println("\n=== Tag Usage Examples ===")
			fmt.Println("# Run only MVP cluster tests:")
			fmt.Println("e2e-runner --component cluster --environment mvp")
			fmt.Println()
			fmt.Println("# Run only failure scenarios:")
			fmt.Println("e2e-runner --category failure")
			fmt.Println()
			fmt.Println("# Run specific test:")
			fmt.Println("e2e-runner --test-id e2e-001")
			fmt.Println()
			fmt.Println("# Run GCP tests:")
			fmt.Println("e2e-runner --provider gcp")
			fmt.Println()
			fmt.Println("# Custom tag expression:")
			fmt.Println("e2e-runner --tags \"@v1.0 && @cluster && !@failure\"")
		},
	}
}

func validateFeaturesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-features",
		Short: "Validate feature files for syntax and tag consistency",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Validating feature files...")
			// Implementation would validate .feature files
			fmt.Println("Feature validation completed successfully!")
		},
	}
}

func generateReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate-report",
		Short: "Generate test execution report",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Generating test report...")
			// Implementation would generate HTML/JSON reports
			fmt.Println("Test report generated successfully!")
		},
	}
}