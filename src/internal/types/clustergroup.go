package types

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// NamespaceEntry represents a namespace that can be either a string or a map with additional configuration
type NamespaceEntry struct {
	value interface{}
}

// MarshalYAML implements the yaml.Marshaler interface for NamespaceEntry
func (ne NamespaceEntry) MarshalYAML() (interface{}, error) {
	return ne.value, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for NamespaceEntry
func (ne *NamespaceEntry) UnmarshalYAML(value *yaml.Node) error {
	ne.value = nil

	var str string
	if err := value.Decode(&str); err == nil {
		ne.value = str
		return nil
	}

	var m map[string]interface{}
	if err := value.Decode(&m); err == nil {
		ne.value = m
		return nil
	}

	return fmt.Errorf("namespaces entry at line %d, column %d must be either a string or a map", value.Line, value.Column)
}

// GetString returns the string value if this NamespaceEntry is a string, and a boolean indicating success
func (ne NamespaceEntry) GetString() (string, bool) {
	if str, ok := ne.value.(string); ok {
		return str, true
	}
	return "", false
}

// Equal compares two NamespaceEntry values for equality
func (ne NamespaceEntry) Equal(other NamespaceEntry) bool {
	// Simple comparison - could be enhanced for deep map comparison if needed
	if str1, ok1 := ne.GetString(); ok1 {
		if str2, ok2 := other.GetString(); ok2 {
			return str1 == str2
		}
	}
	// For maps or other complex types, we'd need deeper comparison
	// For now, assume they're different if not both strings
	return false
}

// NewNamespaceEntry creates a new NamespaceEntry from a string
func NewNamespaceEntry(namespace string) NamespaceEntry {
	return NamespaceEntry{value: namespace}
}

// Application defines the structure for an ArgoCD application entry.
type Application struct {
	Name         string                 `yaml:"name"`
	Namespace    string                 `yaml:"namespace"`
	Project      string                 `yaml:"project"`
	Path         string                 `yaml:"path,omitempty"`
	Chart        string                 `yaml:"chart,omitempty"`
	ChartVersion string                 `yaml:"chartVersion,omitempty"`
	OtherFields  map[string]interface{} `yaml:",inline"`
}

// Subscription defines the structure for an Operator subscription.
type Subscription struct {
	Name        string                 `yaml:"name"`
	Namespace   string                 `yaml:"namespace"`
	Channel     string                 `yaml:"channel,omitempty"`
	Source      string                 `yaml:"source,omitempty"`
	OtherFields map[string]interface{} `yaml:",inline"`
}

// ClusterGroup holds the detailed configuration for the cluster group.
type ClusterGroup struct {
	Name          string                  `yaml:"name"`
	IsHubCluster  bool                    `yaml:"isHubCluster,omitempty"`
	Namespaces    []NamespaceEntry        `yaml:"namespaces"`
	Projects      []string                `yaml:"projects"`
	Subscriptions map[string]Subscription `yaml:"subscriptions"`
	Applications  map[string]Application  `yaml:"applications"`
	OtherFields   map[string]interface{}  `yaml:",inline"`
}

// ValuesClusterGroup is the top-level struct for the cluster group values file.
type ValuesClusterGroup struct {
	ClusterGroup ClusterGroup           `yaml:"clusterGroup"`
	OtherFields  map[string]interface{} `yaml:",inline"`
}

// NewDefaultValuesClusterGroup creates a default configuration for a cluster group.
// It conditionally includes secrets-related resources based on the useSecrets flag.
func NewDefaultValuesClusterGroup(patternName, clusterGroupName string, chartPaths []string, useSecrets bool) *ValuesClusterGroup {
	namespaces := []NamespaceEntry{NewNamespaceEntry(patternName)}
	projects := []string{patternName}
	applications := make(map[string]Application)

	if useSecrets {
		projects = append(projects, clusterGroupName)
	}

	if useSecrets {
		namespaces = append(namespaces, NewNamespaceEntry("vault"), NewNamespaceEntry("golang-external-secrets"))
		applications["vault"] = Application{
			Name:         "vault",
			Namespace:    "vault",
			Project:      clusterGroupName,
			Chart:        "hashicorp-vault",
			ChartVersion: "0.1.*",
		}
		applications["golang-external-secrets"] = Application{
			Name:         "golang-external-secrets",
			Namespace:    "golang-external-secrets",
			Project:      clusterGroupName,
			Chart:        "golang-external-secrets",
			ChartVersion: "0.1.*",
		}
	}

	for _, path := range chartPaths {
		chartName := filepath.Base(path)
		app := Application{
			Name:      chartName,
			Namespace: patternName,
			Project:   patternName,
			Path:      path,
		}
		applications[chartName] = app
	}

	return &ValuesClusterGroup{
		ClusterGroup: ClusterGroup{
			Name:          clusterGroupName,
			Namespaces:    namespaces,
			Projects:      projects,
			Subscriptions: make(map[string]Subscription),
			Applications:  applications,
			OtherFields:   make(map[string]interface{}),
		},
		OtherFields: make(map[string]interface{}),
	}
}
