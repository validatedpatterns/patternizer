package pattern

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/dminnear-rh/patternizer/internal/types"
)

// TestExtractPatternNameFromURL tests URL parsing for different Git URL formats.
func TestExtractPatternNameFromURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:     "SSH URL with .git suffix",
			url:      "git@github.com:user/my-pattern.git",
			expected: "my-pattern",
		},
		{
			name:     "SSH URL without .git suffix",
			url:      "git@github.com:user/my-pattern",
			expected: "my-pattern",
		},
		{
			name:     "HTTPS URL with .git suffix",
			url:      "https://github.com/user/my-pattern.git",
			expected: "my-pattern",
		},
		{
			name:     "HTTPS URL without .git suffix",
			url:      "https://github.com/user/my-pattern",
			expected: "my-pattern",
		},
		{
			name:     "HTTP URL with .git suffix",
			url:      "http://github.com/user/my-pattern.git",
			expected: "my-pattern",
		},
		{
			name:     "HTTP URL without .git suffix",
			url:      "http://github.com/user/my-pattern",
			expected: "my-pattern",
		},
		{
			name:     "GitLab SSH URL",
			url:      "git@gitlab.com:group/subgroup/my-pattern.git",
			expected: "my-pattern",
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/group/subgroup/my-pattern.git",
			expected: "my-pattern",
		},
		{
			name:        "Invalid SSH URL format",
			url:         "git@github.com",
			expectError: true,
		},
		{
			name:        "Unsupported protocol",
			url:         "ftp://github.com/user/repo.git",
			expectError: true,
		},
		{
			name:        "Empty URL",
			url:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractPatternNameFromURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for URL '%s', but got none", tt.url)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for URL '%s': %v", tt.url, err)
				return
			}

			if result != tt.expected {
				t.Errorf("extractPatternNameFromURL('%s') = '%s', expected '%s'", tt.url, result, tt.expected)
			}
		})
	}
}

// TestProcessGlobalValuesPreservesFields tests that ProcessGlobalValues preserves existing user fields.
func TestProcessGlobalValuesPreservesFields(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pattern-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create initial values-global.yaml with custom fields
	initialValues := map[string]interface{}{
		"global": map[string]interface{}{
			"pattern":     "existing-pattern",
			"customField": "customValue",
			"nestedCustom": map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		},
		"main": map[string]interface{}{
			"clusterGroupName": "custom-cluster-group",
			"multiSourceConfig": map[string]interface{}{
				"enabled":                  false,   // Different from default
				"clusterGroupChartVersion": "1.0.*", // Different from default
				"customMultiSource":        "customValue",
			},
			"customMainField": "mainCustomValue",
		},
		"customTopLevel": map[string]interface{}{
			"someKey":    "someValue",
			"anotherKey": []string{"item1", "item2"},
		},
	}

	valuesPath := filepath.Join(tempDir, "values-global.yaml")
	initialYaml, err := yaml.Marshal(initialValues)
	if err != nil {
		t.Fatalf("Failed to marshal initial values: %v", err)
	}
	if err := os.WriteFile(valuesPath, initialYaml, 0o644); err != nil {
		t.Fatalf("Failed to write initial values file: %v", err)
	}

	// Process the values
	actualPatternName, clusterGroupName, err := ProcessGlobalValues("new-pattern", tempDir)
	if err != nil {
		t.Fatalf("ProcessGlobalValues failed: %v", err)
	}

	// Verify return values
	if actualPatternName != "existing-pattern" {
		t.Errorf("Expected pattern name 'existing-pattern', got '%s'", actualPatternName)
	}
	if clusterGroupName != "custom-cluster-group" {
		t.Errorf("Expected cluster group name 'custom-cluster-group', got '%s'", clusterGroupName)
	}

	// Read the processed file
	processedData, err := os.ReadFile(valuesPath)
	if err != nil {
		t.Fatalf("Failed to read processed file: %v", err)
	}

	var processedValues map[string]interface{}
	if err := yaml.Unmarshal(processedData, &processedValues); err != nil {
		t.Fatalf("Failed to unmarshal processed values: %v", err)
	}

	// Verify all custom fields are preserved
	tests := []struct {
		path     []string
		expected interface{}
	}{
		{[]string{"global", "pattern"}, "existing-pattern"},
		{[]string{"global", "customField"}, "customValue"},
		{[]string{"global", "nestedCustom", "key1"}, "value1"},
		{[]string{"global", "nestedCustom", "key2"}, 42},
		{[]string{"main", "clusterGroupName"}, "custom-cluster-group"},
		{[]string{"main", "multiSourceConfig", "enabled"}, false},
		{[]string{"main", "multiSourceConfig", "clusterGroupChartVersion"}, "1.0.*"},
		{[]string{"main", "multiSourceConfig", "customMultiSource"}, "customValue"},
		{[]string{"main", "customMainField"}, "mainCustomValue"},
		{[]string{"customTopLevel", "someKey"}, "someValue"},
	}

	for _, tt := range tests {
		value := getNestedValue(processedValues, tt.path)
		if value != tt.expected {
			t.Errorf("Field %v = %v, expected %v", tt.path, value, tt.expected)
		}
	}

	// Verify array field is preserved
	customTopLevel, ok := processedValues["customTopLevel"].(map[string]interface{})
	if !ok {
		t.Fatalf("customTopLevel is not a map")
	}
	anotherKey, ok := customTopLevel["anotherKey"].([]interface{})
	if !ok {
		t.Fatalf("anotherKey is not an array")
	}
	expectedArray := []interface{}{"item1", "item2"}
	if len(anotherKey) != len(expectedArray) {
		t.Errorf("anotherKey length = %d, expected %d", len(anotherKey), len(expectedArray))
	}
	for i, expected := range expectedArray {
		if i < len(anotherKey) && anotherKey[i] != expected {
			t.Errorf("anotherKey[%d] = %v, expected %v", i, anotherKey[i], expected)
		}
	}
}

// TestProcessClusterGroupValuesPreservesFields tests that ProcessClusterGroupValues preserves existing user fields.
func TestProcessClusterGroupValuesPreservesFields(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pattern-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create initial values-prod.yaml with custom fields
	initialValues := map[string]interface{}{
		"clusterGroup": map[string]interface{}{
			"name":         "prod",
			"isHubCluster": true,
			"namespaces":   []interface{}{"custom-ns1", "custom-ns2"},
			"projects":     []interface{}{"custom-proj1", "custom-proj2"},
			"subscriptions": map[string]interface{}{
				"custom-operator": map[string]interface{}{
					"name":      "custom-operator",
					"namespace": "custom-namespace",
					"channel":   "stable",
					"source":    "community-operators",
				},
			},
			"applications": map[string]interface{}{
				"custom-app": map[string]interface{}{
					"name":           "custom-app",
					"namespace":      "custom-namespace",
					"project":        "custom-project",
					"path":           "custom/path",
					"customAppField": "customAppValue",
				},
			},
			"customClusterField": "customClusterValue",
		},
		"customTopLevel": map[string]interface{}{
			"customKey": "customValue",
		},
	}

	valuesPath := filepath.Join(tempDir, "values-prod.yaml")
	initialYaml, err := yaml.Marshal(initialValues)
	if err != nil {
		t.Fatalf("Failed to marshal initial values: %v", err)
	}
	if err := os.WriteFile(valuesPath, initialYaml, 0o644); err != nil {
		t.Fatalf("Failed to write initial values file: %v", err)
	}

	// Process the values
	chartPaths := []string{"charts/app1", "charts/app2"}
	err = ProcessClusterGroupValues("test-pattern", "prod", tempDir, chartPaths, false)
	if err != nil {
		t.Fatalf("ProcessClusterGroupValues failed: %v", err)
	}

	// Read the processed file
	processedData, err := os.ReadFile(valuesPath)
	if err != nil {
		t.Fatalf("Failed to read processed file: %v", err)
	}

	var processedValues map[string]interface{}
	if err := yaml.Unmarshal(processedData, &processedValues); err != nil {
		t.Fatalf("Failed to unmarshal processed values: %v", err)
	}

	// Verify custom fields are preserved
	tests := []struct {
		path     []string
		expected interface{}
	}{
		{[]string{"clusterGroup", "name"}, "prod"},
		{[]string{"clusterGroup", "isHubCluster"}, true},
		{[]string{"clusterGroup", "customClusterField"}, "customClusterValue"},
		{[]string{"customTopLevel", "customKey"}, "customValue"},
	}

	for _, tt := range tests {
		value := getNestedValue(processedValues, tt.path)
		if value != tt.expected {
			t.Errorf("Field %v = %v, expected %v", tt.path, value, tt.expected)
		}
	}

	// Verify custom application fields are preserved
	clusterGroup, ok := processedValues["clusterGroup"].(map[string]interface{})
	if !ok {
		t.Fatalf("clusterGroup is not a map")
	}
	applications, ok := clusterGroup["applications"].(map[string]interface{})
	if !ok {
		t.Fatalf("applications is not a map")
	}
	customApp, ok := applications["custom-app"].(map[string]interface{})
	if !ok {
		t.Fatalf("custom-app is not a map")
	}
	if customApp["customAppField"] != "customAppValue" {
		t.Errorf("custom-app customAppField = %v, expected 'customAppValue'", customApp["customAppField"])
	}

	// Verify custom subscription is preserved
	subscriptions, ok := clusterGroup["subscriptions"].(map[string]interface{})
	if !ok {
		t.Fatalf("subscriptions is not a map")
	}
	customSub, ok := subscriptions["custom-operator"].(map[string]interface{})
	if !ok {
		t.Fatalf("custom-operator subscription is not a map")
	}
	if customSub["channel"] != "stable" {
		t.Errorf("custom-operator channel = %v, expected 'stable'", customSub["channel"])
	}

	// Verify new applications were added while preserving existing ones
	if _, exists := applications["app1"]; !exists {
		t.Error("Expected new application 'app1' to be added")
	}
	if _, exists := applications["app2"]; !exists {
		t.Error("Expected new application 'app2' to be added")
	}
	if _, exists := applications["custom-app"]; !exists {
		t.Error("Expected existing application 'custom-app' to be preserved")
	}
}

// TestProcessGlobalValuesWithNewFile tests ProcessGlobalValues when no existing file exists.
func TestProcessGlobalValuesWithNewFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pattern-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Process values without existing file
	actualPatternName, clusterGroupName, err := ProcessGlobalValues("test-pattern", tempDir)
	if err != nil {
		t.Fatalf("ProcessGlobalValues failed: %v", err)
	}

	// Verify return values
	if actualPatternName != "test-pattern" {
		t.Errorf("Expected pattern name 'test-pattern', got '%s'", actualPatternName)
	}
	if clusterGroupName != "prod" { // Default cluster group name
		t.Errorf("Expected cluster group name 'prod', got '%s'", clusterGroupName)
	}

	// Verify file was created with defaults
	valuesPath := filepath.Join(tempDir, "values-global.yaml")
	if _, err := os.Stat(valuesPath); os.IsNotExist(err) {
		t.Fatal("values-global.yaml was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(valuesPath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	var values types.ValuesGlobal
	if err := yaml.Unmarshal(data, &values); err != nil {
		t.Fatalf("Failed to unmarshal created file: %v", err)
	}

	if values.Global.Pattern != "test-pattern" {
		t.Errorf("Global pattern = %s, expected 'test-pattern'", values.Global.Pattern)
	}
	if values.Main.ClusterGroupName != "prod" {
		t.Errorf("Main clusterGroupName = %s, expected 'prod'", values.Main.ClusterGroupName)
	}
	if !values.Main.MultiSourceConfig.Enabled {
		t.Error("MultiSourceConfig.Enabled should be true by default")
	}
	if values.Main.MultiSourceConfig.ClusterGroupChartVersion != "0.9.*" {
		t.Errorf("ClusterGroupChartVersion = %s, expected '0.9.*'", values.Main.MultiSourceConfig.ClusterGroupChartVersion)
	}
}

// getNestedValue is a helper function to get nested values from a map using a path.
func getNestedValue(m map[string]interface{}, path []string) interface{} {
	current := m
	for i, key := range path {
		if i == len(path)-1 {
			return current[key]
		}
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return nil
		}
		current = next
	}
	return nil
}
