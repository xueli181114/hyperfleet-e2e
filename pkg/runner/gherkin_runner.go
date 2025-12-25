package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"hyperfleet-e2e/pkg/tags"
)

// GherkinRunner integrates Godog (Gherkin) with Ginkgo/Gomega
type GherkinRunner struct {
	tagParser    *tags.TagParser
	stepRegistry *StepRegistry
	config       *RunnerConfig
	suiteContext *SuiteContext
}

// RunnerConfig holds configuration for the test runner
type RunnerConfig struct {
	FeaturePath      string
	OutputFormat     string
	Randomize        bool
	StopOnFailure    bool
	Strict           bool
	NoColors         bool
	Tags             string
	Concurrency      int
	Timeout          time.Duration
	APIBaseURL       string
	AuthToken        string
	DebugMode        bool
}

// SuiteContext holds shared context across scenarios
type SuiteContext struct {
	APIClient    *APIClient
	TestData     map[string]interface{}
	Cleanup      []func() error
	CurrentTags  *tags.VersionTag
	ScenarioName string
}

// NewGherkinRunner creates a new Gherkin runner instance
func NewGherkinRunner(config *RunnerConfig) *GherkinRunner {
	return &GherkinRunner{
		tagParser:    tags.NewTagParser(),
		stepRegistry: NewStepRegistry(),
		config:       config,
		suiteContext: &SuiteContext{
			TestData: make(map[string]interface{}),
			Cleanup:  make([]func() error, 0),
		},
	}
}

// RunFeatures executes Gherkin features with Ginkgo integration
func (gr *GherkinRunner) RunFeatures() {
	// Initialize Gomega for assertions
	gomega.RegisterFailHandler(ginkgo.Fail)

	// Describe the test suite
	ginkgo.Describe("HyperFleet E2E Tests", func() {
		var suiteCtx *SuiteContext

		ginkgo.BeforeSuite(func() {
			suiteCtx = gr.setupSuite()
		})

		ginkgo.AfterSuite(func() {
			gr.cleanupSuite(suiteCtx)
		})

		// Dynamically generate tests from Gherkin features
		gr.generateGinkgoSpecs()
	})
}

// generateGinkgoSpecs creates Ginkgo specs from Gherkin scenarios
func (gr *GherkinRunner) generateGinkgoSpecs() {
	features := gr.loadFeatureFiles()

	for _, feature := range features {
		// Create a Describe block for each feature
		ginkgo.Describe(fmt.Sprintf("Feature: %s", feature.Name), func() {

			for _, scenario := range feature.Scenarios {
				// Parse tags for this scenario
				versionTag := gr.tagParser.ParseTags(scenario.Tags)
				ginkgoLabels := versionTag.ToGinkgoLabels()

				// Create an It block for each scenario with automatic labels
				ginkgo.It(scenario.Name, ginkgo.Label(ginkgoLabels...), func(ctx ginkgo.SpecContext) {
					gr.executeScenario(ctx, &scenario, versionTag)
				})
			}
		})
	}
}

// executeScenario runs a single Gherkin scenario
func (gr *GherkinRunner) executeScenario(ctx ginkgo.SpecContext, scenario *Scenario, versionTag *tags.VersionTag) {
	// Set up scenario context
	scenarioCtx := gr.setupScenarioContext(scenario, versionTag)
	defer gr.cleanupScenario(scenarioCtx)

	// Execute scenario steps
	for _, step := range scenario.Steps {
		ginkgo.By(fmt.Sprintf("%s %s", step.Keyword, step.Text), func() {
			gr.executeStep(scenarioCtx, &step)
		})
	}
}

// executeStep runs a single Gherkin step
func (gr *GherkinRunner) executeStep(ctx *SuiteContext, step *Step) {
	stepFunc := gr.stepRegistry.GetStepFunction(step.Text)
	if stepFunc == nil {
		ginkgo.Fail(fmt.Sprintf("No step definition found for: %s", step.Text))
		return
	}

	// Execute step with context
	err := stepFunc(ctx, step)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Step failed: %s - Error: %v", step.Text, err))
	}
}

// setupSuite initializes the test suite
func (gr *GherkinRunner) setupSuite() *SuiteContext {
	ginkgo.By("Setting up test suite")

	// Initialize API client
	apiClient := NewAPIClient(gr.config.APIBaseURL, gr.config.AuthToken)

	gr.suiteContext.APIClient = apiClient

	// Verify API connectivity
	err := apiClient.HealthCheck()
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "API should be accessible")

	return gr.suiteContext
}

// cleanupSuite cleans up after the test suite
func (gr *GherkinRunner) cleanupSuite(ctx *SuiteContext) {
	ginkgo.By("Cleaning up test suite")

	// Execute cleanup functions in reverse order
	for i := len(ctx.Cleanup) - 1; i >= 0; i-- {
		if err := ctx.Cleanup[i](); err != nil {
			ginkgo.GinkgoLogr.Error(err, "Failed to execute cleanup function")
		}
	}
}

// setupScenarioContext prepares context for a scenario
func (gr *GherkinRunner) setupScenarioContext(scenario *Scenario, versionTag *tags.VersionTag) *SuiteContext {
	scenarioCtx := &SuiteContext{
		APIClient:    gr.suiteContext.APIClient,
		TestData:     make(map[string]interface{}),
		Cleanup:      make([]func() error, 0),
		CurrentTags:  versionTag,
		ScenarioName: scenario.Name,
	}

	// Add version-specific setup if needed
	if versionTag.Version != "" {
		scenarioCtx.TestData["version"] = versionTag.Version
	}
	if versionTag.Component != "" {
		scenarioCtx.TestData["component"] = versionTag.Component
	}

	return scenarioCtx
}

// cleanupScenario cleans up after a scenario
func (gr *GherkinRunner) cleanupScenario(ctx *SuiteContext) {
	// Execute scenario-specific cleanup
	for i := len(ctx.Cleanup) - 1; i >= 0; i-- {
		if err := ctx.Cleanup[i](); err != nil {
			ginkgo.GinkgoLogr.Error(err, "Failed to execute scenario cleanup")
		}
	}
}

// RunWithGodog runs tests using the traditional Godog approach
func (gr *GherkinRunner) RunWithGodog() int {
	opts := godog.Options{
		Format:        gr.config.OutputFormat,
		Paths:         []string{gr.config.FeaturePath},
		Randomize:     gr.config.Randomize,
		StopOnFailure: gr.config.StopOnFailure,
		Strict:        gr.config.Strict,
		NoColors:      gr.config.NoColors,
		Tags:          gr.config.Tags,
		Concurrency:   gr.config.Concurrency,
	}

	if gr.config.NoColors {
		opts.Output = colors.Colored(colors.None)
	}

	suite := godog.TestSuite{
		Name:                 "HyperFleet E2E Tests",
		Options:              &opts,
		ScenarioInitializer:  gr.initializeScenario,
		TestSuiteInitializer: gr.initializeTestSuite,
	}

	return suite.Run()
}

// initializeTestSuite initializes the test suite for Godog
func (gr *GherkinRunner) initializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		gr.suiteContext = gr.setupSuite()
	})

	ctx.AfterSuite(func() {
		gr.cleanupSuite(gr.suiteContext)
	})
}

// initializeScenario initializes individual scenarios for Godog
func (gr *GherkinRunner) initializeScenario(ctx *godog.ScenarioContext) {
	// Register step definitions
	gr.stepRegistry.RegisterSteps(ctx)

	// Scenario hooks
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		versionTag := gr.tagParser.ParseTags(sc.Tags)
		scenarioCtx := gr.setupScenarioContext(&Scenario{
			Name: sc.Name,
			Tags: sc.Tags,
		}, versionTag)

		// Store scenario context in Godog context
		return context.WithValue(ctx, "scenarioContext", scenarioCtx), nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		scenarioCtx := ctx.Value("scenarioContext").(*SuiteContext)
		gr.cleanupScenario(scenarioCtx)
		return ctx, nil
	})
}

// Feature and Scenario structures for parsing
type Feature struct {
	Name      string
	Scenarios []Scenario
	Tags      []string
}

type Scenario struct {
	Name string
	Tags []string
	Steps []Step
}

type Step struct {
	Keyword string
	Text    string
	Table   *DataTable
	DocString string
}

type DataTable struct {
	Rows [][]string
}

// loadFeatureFiles loads and parses Gherkin feature files
func (gr *GherkinRunner) loadFeatureFiles() []Feature {
	// This would typically parse .feature files from the filesystem
	// For this example, we'll return a simplified structure
	features := make([]Feature, 0)

	featureFiles, _ := filepath.Glob(filepath.Join(gr.config.FeaturePath, "*.feature"))

	for _, file := range featureFiles {
		// Parse feature file (simplified for this example)
		// In a real implementation, you'd use a proper Gherkin parser
		feature := Feature{
			Name: filepath.Base(file),
		}
		features = append(features, feature)
	}

	return features
}