package main

import (
	"os"
	"testing"

	"github.com/dminnear-rh/patternizer/internal/fileutils"
	"github.com/dminnear-rh/patternizer/internal/types"
)

func TestGetResourcePath(t *testing.T) {
	// Test with environment variable set
	os.Setenv("PATTERNIZER_RESOURCES_DIR", "/tmp/test")
	path, err := fileutils.GetResourcePath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != "/tmp/test" {
		t.Fatalf("Expected /tmp/test, got %s", path)
	}

	// Test with environment variable unset
	os.Unsetenv("PATTERNIZER_RESOURCES_DIR")
	path, err = fileutils.GetResourcePath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Should return current directory
	if path == "" {
		t.Fatalf("Expected non-empty path")
	}
}

func TestNewDefaultValuesGlobal(t *testing.T) {
	values := types.NewDefaultValuesGlobal()

	if values.Main.ClusterGroupName != "prod" {
		t.Errorf("Expected clusterGroupName to be 'prod', got '%s'", values.Main.ClusterGroupName)
	}

	if !values.Main.MultiSourceConfig.Enabled {
		t.Error("Expected multiSourceConfig.enabled to be true")
	}

	if values.Main.MultiSourceConfig.ClusterGroupChartVersion != "0.9.*" {
		t.Errorf("Expected clusterGroupChartVersion to be '0.9.*', got '%s'", values.Main.MultiSourceConfig.ClusterGroupChartVersion)
	}
}

func TestNewDefaultValuesClusterGroup(t *testing.T) {
	// Test without secrets
	values := types.NewDefaultValuesClusterGroup("test-pattern", "test-group", []string{"charts/app1", "charts/app2"}, false)

	if values.ClusterGroup.Name != "test-group" {
		t.Errorf("Expected name to be 'test-group', got '%s'", values.ClusterGroup.Name)
	}

	expectedNamespaces := []string{"test-pattern"}
	if len(values.ClusterGroup.Namespaces) != len(expectedNamespaces) {
		t.Errorf("Expected %d namespaces, got %d", len(expectedNamespaces), len(values.ClusterGroup.Namespaces))
	}

	expectedProjects := []string{"test-pattern"}
	if len(values.ClusterGroup.Projects) != len(expectedProjects) {
		t.Errorf("Expected %d projects, got %d", len(expectedProjects), len(values.ClusterGroup.Projects))
	}

	// Test with secrets
	valuesWithSecrets := types.NewDefaultValuesClusterGroup("test-pattern", "test-group", []string{"charts/app1"}, true)

	expectedNamespacesWithSecrets := []string{"test-pattern", "vault", "golang-external-secrets"}
	if len(valuesWithSecrets.ClusterGroup.Namespaces) != len(expectedNamespacesWithSecrets) {
		t.Errorf("Expected %d namespaces with secrets, got %d", len(expectedNamespacesWithSecrets), len(valuesWithSecrets.ClusterGroup.Namespaces))
	}

	// Check that vault and golang-external-secrets applications are added
	if _, exists := valuesWithSecrets.ClusterGroup.Applications["vault"]; !exists {
		t.Error("Expected vault application to be present with secrets")
	}

	if _, exists := valuesWithSecrets.ClusterGroup.Applications["golang-external-secrets"]; !exists {
		t.Error("Expected golang-external-secrets application to be present with secrets")
	}
}
