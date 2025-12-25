package tags

import (
	"fmt"
	"regexp"
	"strings"
)

// VersionTag represents a version control tag with semantic meaning
type VersionTag struct {
	Version     string   // e.g., "v1.0", "v1.1"
	Component   string   // e.g., "cluster", "nodepool", "adapter"
	Action      string   // e.g., "create", "update", "delete"
	Category    string   // e.g., "happy-path", "failure", "validation"
	TestID      string   // e.g., "e2e-001", "e2e-fail-002"
	Provider    string   // e.g., "gcp", "aws", "azure"
	Environment string   // e.g., "mvp", "post-mvp"
	Labels      []string // Additional labels for filtering
}

// TagParser handles parsing and conversion of Gherkin tags
type TagParser struct {
	versionPattern *regexp.Regexp
	testIDPattern  *regexp.Regexp
}

// NewTagParser creates a new tag parser instance
func NewTagParser() *TagParser {
	return &TagParser{
		versionPattern: regexp.MustCompile(`^v\d+\.\d+$`),
		testIDPattern:  regexp.MustCompile(`^e2e(-\w+)?-\d+$`),
	}
}

// ParseTags extracts structured information from Gherkin scenario tags
func (p *TagParser) ParseTags(gherkinTags []string) *VersionTag {
	vt := &VersionTag{
		Labels: make([]string, 0),
	}

	for _, tag := range gherkinTags {
		tag = strings.TrimPrefix(tag, "@")

		switch {
		case p.versionPattern.MatchString(tag):
			vt.Version = tag
		case p.testIDPattern.MatchString(tag):
			vt.TestID = tag
		case p.isComponent(tag):
			vt.Component = tag
		case p.isAction(tag):
			vt.Action = tag
		case p.isCategory(tag):
			vt.Category = tag
		case p.isProvider(tag):
			vt.Provider = tag
		case p.isEnvironment(tag):
			vt.Environment = tag
		default:
			vt.Labels = append(vt.Labels, tag)
		}
	}

	return vt
}

// ToGinkgoLabels converts VersionTag to Ginkgo label expressions
func (vt *VersionTag) ToGinkgoLabels() []string {
	labels := make([]string, 0)

	// Add version labels
	if vt.Version != "" {
		labels = append(labels, vt.Version)
		// Add major version for compatibility
		majorVersion := strings.Split(vt.Version, ".")[0]
		if majorVersion != vt.Version {
			labels = append(labels, majorVersion)
		}
	}

	// Add component labels
	if vt.Component != "" {
		labels = append(labels, vt.Component)
	}

	// Add action labels
	if vt.Action != "" {
		labels = append(labels, vt.Action)
		labels = append(labels, fmt.Sprintf("%s-%s", vt.Component, vt.Action))
	}

	// Add category labels
	if vt.Category != "" {
		labels = append(labels, vt.Category)
	}

	// Add test ID labels
	if vt.TestID != "" {
		labels = append(labels, vt.TestID)
	}

	// Add provider labels
	if vt.Provider != "" {
		labels = append(labels, vt.Provider)
		labels = append(labels, fmt.Sprintf("provider-%s", vt.Provider))
	}

	// Add environment labels
	if vt.Environment != "" {
		labels = append(labels, vt.Environment)
	}

	// Add additional labels
	labels = append(labels, vt.Labels...)

	return labels
}

// GenerateLabelSelector creates Ginkgo label selector expressions
func (vt *VersionTag) GenerateLabelSelector(filters map[string]string) string {
	var expressions []string

	for key, value := range filters {
		switch key {
		case "version":
			if value != "" {
				expressions = append(expressions, value)
			}
		case "component":
			if value != "" {
				expressions = append(expressions, value)
			}
		case "action":
			if value != "" && vt.Component != "" {
				expressions = append(expressions, fmt.Sprintf("%s-%s", vt.Component, value))
			}
		case "category":
			if value != "" {
				expressions = append(expressions, value)
			}
		case "provider":
			if value != "" {
				expressions = append(expressions, fmt.Sprintf("provider-%s", value))
			}
		case "testid":
			if value != "" {
				expressions = append(expressions, value)
			}
		}
	}

	if len(expressions) == 0 {
		return ""
	}

	return strings.Join(expressions, " && ")
}

// Helper functions to categorize tags
func (p *TagParser) isComponent(tag string) bool {
	components := []string{"cluster", "nodepool", "adapter", "api", "sentinel"}
	return p.contains(components, tag)
}

func (p *TagParser) isAction(tag string) bool {
	actions := []string{"create", "update", "delete", "list", "get", "validation"}
	return p.contains(actions, tag)
}

func (p *TagParser) isCategory(tag string) bool {
	categories := []string{
		"happy-path", "failure", "validation", "business-logic-failure",
		"infrastructure-failure", "execution-failure", "lifecycle",
	}
	return p.contains(categories, tag)
}

func (p *TagParser) isProvider(tag string) bool {
	providers := []string{"gcp", "aws", "azure", "openstack"}
	return p.contains(providers, tag)
}

func (p *TagParser) isEnvironment(tag string) bool {
	environments := []string{"mvp", "post-mvp", "alpha", "beta", "production"}
	return p.contains(environments, tag)
}

func (p *TagParser) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetAvailableFilters returns all available filter options
func GetAvailableFilters() map[string][]string {
	return map[string][]string{
		"version":   {"v1.0", "v1.1", "v2.0"},
		"component": {"cluster", "nodepool", "adapter", "api", "sentinel"},
		"action":    {"create", "update", "delete", "list", "get", "validation"},
		"category":  {"happy-path", "failure", "validation", "lifecycle"},
		"provider":  {"gcp", "aws", "azure", "openstack"},
		"environment": {"mvp", "post-mvp", "alpha", "beta", "production"},
	}
}