package main

import "path/filepath"

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
	IsHubCluster  bool                    `yaml:"isHubCluster"`
	Namespaces    []string                `yaml:"namespaces"`
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

// newDefaultValuesClusterGroup creates a default configuration for a cluster group.
// It conditionally includes secrets-related resources based on the useSecrets flag.
func newDefaultValuesClusterGroup(patternName, clusterGroupName string, chartPaths []string, useSecrets bool) *ValuesClusterGroup {
	namespaces := []string{patternName}
	projects := []string{clusterGroupName, patternName}
	applications := make(map[string]Application)

	if useSecrets {
		namespaces = append(namespaces, "vault", "golang-external-secrets")
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
			IsHubCluster:  true,
			Namespaces:    namespaces,
			Projects:      projects,
			Subscriptions: make(map[string]Subscription),
			Applications:  applications,
		},
	}
}
