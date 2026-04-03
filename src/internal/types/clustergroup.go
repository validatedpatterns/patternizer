package types

import (
	"fmt"
	"path/filepath"
	"reflect"

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
	return reflect.DeepEqual(ne.value, other.value)
}

// NewNamespaceEntry creates a new NamespaceEntry from a string
func NewNamespaceEntry(namespace string) NamespaceEntry {
	return NamespaceEntry{value: namespace}
}

// NewMapNamespaceEntry creates a new NamespaceEntry from a map
func NewMapNamespaceEntry(m map[string]interface{}) NamespaceEntry {
	return NamespaceEntry{value: m}
}

// Application defines the structure for an ArgoCD application entry.
type Application struct {
	Name         string                 `yaml:"name"`
	Namespace    string                 `yaml:"namespace"`
	Project      string                 `yaml:"project,omitempty"`
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
	Namespaces    []NamespaceEntry        `yaml:"namespaces"`
	Projects      []string                `yaml:"projects,omitempty"`
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
	applications := make(map[string]Application)
	subscriptions := make(map[string]Subscription)

	if useSecrets {
		namespaces = append(
			namespaces,
			NewNamespaceEntry("vault"),
			NamespaceEntry{map[string]interface{}{
				"external-secrets-operator": map[string]interface{}{
					"operatorGroup":    true,
					"targetNamespaces": []string{},
				},
			},
			},
			NewNamespaceEntry("external-secrets"),
		)
		subscriptions["eso"] = Subscription{
			Name:      "openshift-external-secrets-operator",
			Namespace: "external-secrets-operator",
			Channel:   "stable-v1",
		}
		applications["vault"] = Application{
			Name:         "vault",
			Namespace:    "vault",
			Chart:        "hashicorp-vault",
			ChartVersion: "0.1.*",
		}
		applications["openshift-external-secrets"] = Application{
			Name:         "openshift-external-secrets",
			Namespace:    "external-secrets",
			Chart:        "openshift-external-secrets",
			ChartVersion: "0.0.*",
		}
	}

	for _, path := range chartPaths {
		chartName := filepath.Base(path)
		app := Application{
			Name:      chartName,
			Namespace: patternName,
			Path:      path,
		}
		applications[chartName] = app
	}

	return &ValuesClusterGroup{
		ClusterGroup: ClusterGroup{
			Name:          clusterGroupName,
			Namespaces:    namespaces,
			Subscriptions: subscriptions,
			Applications:  applications,
			OtherFields:   make(map[string]interface{}),
		},
		OtherFields: make(map[string]interface{}),
	}
}
