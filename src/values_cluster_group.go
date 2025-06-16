package main

import "path/filepath"

// Application defines the structure for an ArgoCD application entry.
type Application struct {
	Name         string `yaml:"name"`
	Namespace    string `yaml:"namespace"`
	Project      string `yaml:"project"`
	Path         string `yaml:"path,omitempty"`
	Chart        string `yaml:"chart,omitempty"`
	ChartVersion string `yaml:"chartVersion,omitempty"`
}

// Subscription defines the structure for an Operator subscription.
type Subscription struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Channel   string `yaml:"channel,omitempty"`
	Source    string `yaml:"source,omitempty"`
}

// ClusterGroup holds the detailed configuration for the cluster group.
type ClusterGroup struct {
	Name          string                  `yaml:"name"`
	IsHubCluster  bool                    `yaml:"isHubCluster"`
	Namespaces    []string                `yaml:"namespaces"`
	Projects      []string                `yaml:"projects"`
	Subscriptions map[string]Subscription `yaml:"subscriptions"`
	Applications  map[string]Application  `yaml:"applications"`
}

// ValuesClusterGroup is the top-level struct for the cluster group values file.
type ValuesClusterGroup struct {
	ClusterGroup ClusterGroup `yaml:"clusterGroup"`
}

// newDefaultValuesClusterGroup creates a default configuration for a cluster group.
func newDefaultValuesClusterGroup(patternName, clusterGroupName string, chartPaths []string) *ValuesClusterGroup {
	// Initialize with default namespaces and projects
	namespaces := []string{"vault", "golang-external-secrets", patternName}
	projects := []string{clusterGroupName, patternName}

	// Initialize with default applications that are always present
	applications := map[string]Application{
		"vault": {
			Name:         "vault",
			Namespace:    "vault",
			Project:      clusterGroupName,
			Chart:        "hashicorp-vault",
			ChartVersion: "0.1.*",
		},
		"golang-external-secrets": {
			Name:         "golang-external-secrets",
			Namespace:    "golang-external-secrets",
			Project:      clusterGroupName,
			Chart:        "golang-external-secrets",
			ChartVersion: "0.1.*",
		},
	}

	// Add applications discovered from the filesystem
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
