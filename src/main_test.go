package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetResourcePath(t *testing.T) {
	// Test with environment variable set
	os.Setenv("PATTERNIZER_RESOURCES_DIR", "/test/resources")
	result := getResourcePath("test.yaml")
	expected := "/test/resources/test.yaml"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test without environment variable (fallback)
	os.Unsetenv("PATTERNIZER_RESOURCES_DIR")
	result = getResourcePath("test.yaml")
	expected = "test.yaml"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestNewDefaultValuesGlobal(t *testing.T) {
	values := newDefaultValuesGlobal()

	if values == nil {
		t.Fatal("Expected non-nil ValuesGlobal")
	}

	if values.Main.ClusterGroupName != "prod" {
		t.Errorf("Expected default cluster group name 'prod', got '%s'", values.Main.ClusterGroupName)
	}

	if !values.Main.MultiSourceConfig.Enabled {
		t.Error("Expected MultiSourceConfig.Enabled to be true")
	}

	if values.Main.MultiSourceConfig.ClusterGroupChartVersion != "0.9.*" {
		t.Errorf("Expected chart version '0.9.*', got '%s'", values.Main.MultiSourceConfig.ClusterGroupChartVersion)
	}
}

func TestNewDefaultValuesClusterGroup(t *testing.T) {
	patternName := "test-pattern"
	clusterGroupName := "test-cluster"
	chartPaths := []string{"charts/app1", "charts/app2"}

	values := newDefaultValuesClusterGroup(patternName, clusterGroupName, chartPaths, false)

	if values == nil {
		t.Fatal("Expected non-nil ValuesClusterGroup")
	}

	if values.ClusterGroup.Name != clusterGroupName {
		t.Errorf("Expected cluster group name '%s', got '%s'", clusterGroupName, values.ClusterGroup.Name)
	}

	expectedNamespaces := []string{patternName}
	if len(values.ClusterGroup.Namespaces) != len(expectedNamespaces) {
		t.Errorf("Expected %d namespaces, got %d", len(expectedNamespaces), len(values.ClusterGroup.Namespaces))
	}

	expectedProjects := []string{clusterGroupName, patternName}
	if len(values.ClusterGroup.Projects) != len(expectedProjects) {
		t.Errorf("Expected %d projects, got %d", len(expectedProjects), len(values.ClusterGroup.Projects))
	}

	if len(values.ClusterGroup.Applications) != len(chartPaths) {
		t.Errorf("Expected %d applications, got %d", len(chartPaths), len(values.ClusterGroup.Applications))
	}

	// Check that applications are created correctly
	for _, chartPath := range chartPaths {
		chartName := filepath.Base(chartPath)
		app, exists := values.ClusterGroup.Applications[chartName]
		if !exists {
			t.Errorf("Expected application '%s' to exist", chartName)
			continue
		}

		if app.Name != chartName {
			t.Errorf("Expected app name '%s', got '%s'", chartName, app.Name)
		}

		if app.Path != chartPath {
			t.Errorf("Expected app path '%s', got '%s'", chartPath, app.Path)
		}

		if app.Namespace != patternName {
			t.Errorf("Expected app namespace '%s', got '%s'", patternName, app.Namespace)
		}

		if app.Project != patternName {
			t.Errorf("Expected app project '%s', got '%s'", patternName, app.Project)
		}
	}
}

func TestNewDefaultValuesClusterGroupWithSecrets(t *testing.T) {
	patternName := "test-pattern"
	clusterGroupName := "test-cluster"
	chartPaths := []string{"charts/app1"}

	values := newDefaultValuesClusterGroup(patternName, clusterGroupName, chartPaths, true)

	// Should have additional namespaces for secrets
	expectedNamespaces := []string{patternName, "vault", "golang-external-secrets"}
	if len(values.ClusterGroup.Namespaces) != len(expectedNamespaces) {
		t.Errorf("Expected %d namespaces with secrets, got %d", len(expectedNamespaces), len(values.ClusterGroup.Namespaces))
	}

	// Should have vault and golang-external-secrets applications
	if _, exists := values.ClusterGroup.Applications["vault"]; !exists {
		t.Error("Expected vault application to exist when secrets enabled")
	}

	if _, exists := values.ClusterGroup.Applications["golang-external-secrets"]; !exists {
		t.Error("Expected golang-external-secrets application to exist when secrets enabled")
	}
}
