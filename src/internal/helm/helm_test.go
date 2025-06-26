package helm

import (
	"os"
	"path/filepath"
	"testing"
)

// createTestChartStructure creates a comprehensive test directory structure for helm chart testing.
// It returns the temp directory path that should be cleaned up by the caller.
func createTestChartStructure(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "helm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test directory structure:
	// tempDir/
	//   ├── chart1/                    (valid top-level chart)
	//   │   ├── Chart.yaml
	//   │   ├── values.yaml
	//   │   └── templates/
	//   ├── chart2/                    (valid top-level chart)
	//   │   ├── Chart.yaml
	//   │   ├── values.yaml
	//   │   ├── templates/
	//   │   └── charts/                (sub-chart directory)
	//   │       └── subchart/          (sub-chart - should be ignored)
	//   │           ├── Chart.yaml
	//   │           ├── values.yaml
	//   │           └── templates/
	//   ├── incomplete-chart/          (invalid chart - missing templates)
	//   │   ├── Chart.yaml
	//   │   └── values.yaml
	//   ├── missing-chart-yaml/        (invalid chart - missing Chart.yaml)
	//   │   ├── values.yaml
	//   │   └── templates/
	//   ├── missing-values-yaml/       (invalid chart - missing values.yaml)
	//   │   ├── Chart.yaml
	//   │   └── templates/
	//   ├── templates-is-file/         (invalid chart - templates is a file)
	//   │   ├── Chart.yaml
	//   │   ├── values.yaml
	//   │   └── templates              (file, not directory)
	//   ├── .hidden-chart/             (hidden directory - should be ignored)
	//   │   ├── Chart.yaml
	//   │   ├── values.yaml
	//   │   └── templates/
	//   └── not-a-chart/               (not a chart directory)
	//       └── some-file.txt

	// Create chart1 (valid top-level chart)
	chart1Dir := filepath.Join(tempDir, "chart1")
	if err := os.MkdirAll(filepath.Join(chart1Dir, "templates"), 0o755); err != nil {
		t.Fatalf("Failed to create chart1 structure: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chart1Dir, "Chart.yaml"), []byte("name: chart1\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create Chart.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chart1Dir, "values.yaml"), []byte("# values\n"), 0o644); err != nil {
		t.Fatalf("Failed to create values.yaml: %v", err)
	}

	// Create chart2 (valid top-level chart with sub-chart)
	chart2Dir := filepath.Join(tempDir, "chart2")
	chart2TemplatesDir := filepath.Join(chart2Dir, "templates")
	chart2SubchartDir := filepath.Join(chart2Dir, "charts", "subchart")
	subchartTemplatesDir := filepath.Join(chart2SubchartDir, "templates")

	if err := os.MkdirAll(chart2TemplatesDir, 0o755); err != nil {
		t.Fatalf("Failed to create chart2 structure: %v", err)
	}
	if err := os.MkdirAll(subchartTemplatesDir, 0o755); err != nil {
		t.Fatalf("Failed to create subchart structure: %v", err)
	}

	// Chart2 files
	if err := os.WriteFile(filepath.Join(chart2Dir, "Chart.yaml"), []byte("name: chart2\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create chart2 Chart.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chart2Dir, "values.yaml"), []byte("# values\n"), 0o644); err != nil {
		t.Fatalf("Failed to create chart2 values.yaml: %v", err)
	}

	// Sub-chart files (should be ignored by FindTopLevelCharts)
	if err := os.WriteFile(filepath.Join(chart2SubchartDir, "Chart.yaml"), []byte("name: subchart\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create subchart Chart.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chart2SubchartDir, "values.yaml"), []byte("# subchart values\n"), 0o644); err != nil {
		t.Fatalf("Failed to create subchart values.yaml: %v", err)
	}

	// Create incomplete-chart (missing templates directory)
	incompleteChartDir := filepath.Join(tempDir, "incomplete-chart")
	if err := os.MkdirAll(incompleteChartDir, 0o755); err != nil {
		t.Fatalf("Failed to create incomplete-chart structure: %v", err)
	}
	if err := os.WriteFile(filepath.Join(incompleteChartDir, "Chart.yaml"), []byte("name: incomplete\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create incomplete Chart.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(incompleteChartDir, "values.yaml"), []byte("# values\n"), 0o644); err != nil {
		t.Fatalf("Failed to create incomplete values.yaml: %v", err)
	}

	// Create missing-chart-yaml (missing Chart.yaml)
	missingChartYamlDir := filepath.Join(tempDir, "missing-chart-yaml")
	if err := os.MkdirAll(filepath.Join(missingChartYamlDir, "templates"), 0o755); err != nil {
		t.Fatalf("Failed to create missing-chart-yaml structure: %v", err)
	}
	if err := os.WriteFile(filepath.Join(missingChartYamlDir, "values.yaml"), []byte("# values\n"), 0o644); err != nil {
		t.Fatalf("Failed to create missing-chart-yaml values.yaml: %v", err)
	}

	// Create missing-values-yaml (missing values.yaml)
	missingValuesYamlDir := filepath.Join(tempDir, "missing-values-yaml")
	if err := os.MkdirAll(filepath.Join(missingValuesYamlDir, "templates"), 0o755); err != nil {
		t.Fatalf("Failed to create missing-values-yaml structure: %v", err)
	}
	if err := os.WriteFile(filepath.Join(missingValuesYamlDir, "Chart.yaml"), []byte("name: missing-values\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create missing-values-yaml Chart.yaml: %v", err)
	}

	// Create templates-is-file (templates is a file, not directory)
	templatesIsFileDir := filepath.Join(tempDir, "templates-is-file")
	if err := os.MkdirAll(templatesIsFileDir, 0o755); err != nil {
		t.Fatalf("Failed to create templates-is-file structure: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templatesIsFileDir, "Chart.yaml"), []byte("name: templates-file\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create templates-is-file Chart.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templatesIsFileDir, "values.yaml"), []byte("# values\n"), 0o644); err != nil {
		t.Fatalf("Failed to create templates-is-file values.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templatesIsFileDir, "templates"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("Failed to create templates-is-file templates file: %v", err)
	}

	// Create hidden chart (should be ignored by FindTopLevelCharts)
	hiddenChartDir := filepath.Join(tempDir, ".hidden-chart")
	if err := os.MkdirAll(filepath.Join(hiddenChartDir, "templates"), 0o755); err != nil {
		t.Fatalf("Failed to create hidden chart structure: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenChartDir, "Chart.yaml"), []byte("name: hidden\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("Failed to create hidden Chart.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenChartDir, "values.yaml"), []byte("# values\n"), 0o644); err != nil {
		t.Fatalf("Failed to create hidden values.yaml: %v", err)
	}

	// Create not-a-chart directory
	notChartDir := filepath.Join(tempDir, "not-a-chart")
	if err := os.MkdirAll(notChartDir, 0o755); err != nil {
		t.Fatalf("Failed to create not-a-chart structure: %v", err)
	}
	if err := os.WriteFile(filepath.Join(notChartDir, "some-file.txt"), []byte("not a chart\n"), 0o644); err != nil {
		t.Fatalf("Failed to create some-file.txt: %v", err)
	}

	return tempDir
}

// TestFindTopLevelCharts tests that FindTopLevelCharts only returns top-level charts
// and properly skips sub-charts and non-chart directories.
func TestFindTopLevelCharts(t *testing.T) {
	tempDir := createTestChartStructure(t)
	defer os.RemoveAll(tempDir)

	// Run FindTopLevelCharts
	charts, err := FindTopLevelCharts(tempDir)
	if err != nil {
		t.Fatalf("FindTopLevelCharts failed: %v", err)
	}

	// Verify results
	expectedCharts := []string{"chart1", "chart2"}
	if len(charts) != len(expectedCharts) {
		t.Fatalf("Expected %d charts, got %d: %v", len(expectedCharts), len(charts), charts)
	}

	// Convert to map for easier checking
	foundCharts := make(map[string]bool)
	for _, chart := range charts {
		foundCharts[chart] = true
	}

	// Verify each expected chart was found
	for _, expected := range expectedCharts {
		if !foundCharts[expected] {
			t.Errorf("Expected chart '%s' not found in results: %v", expected, charts)
		}
	}

	// Verify that sub-charts, incomplete charts, and hidden charts were NOT found
	unexpectedCharts := []string{"charts/subchart", "subchart", "incomplete-chart", ".hidden-chart", "not-a-chart"}
	for _, unexpected := range unexpectedCharts {
		if foundCharts[unexpected] {
			t.Errorf("Unexpected chart '%s' found in results: %v", unexpected, charts)
		}
	}
}

// TestIsHelmChart tests the IsHelmChart function with various directory structures.
func TestIsHelmChart(t *testing.T) {
	tempDir := createTestChartStructure(t)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		chartDir       string
		expectedResult bool
	}{
		{
			name:           "valid helm chart 1",
			chartDir:       "chart1",
			expectedResult: true,
		},
		{
			name:           "valid helm chart 2",
			chartDir:       "chart2",
			expectedResult: true,
		},
		{
			name:           "valid subchart (still valid helm chart)",
			chartDir:       "chart2/charts/subchart",
			expectedResult: true,
		},
		{
			name:           "hidden chart (still valid helm chart)",
			chartDir:       ".hidden-chart",
			expectedResult: true,
		},
		{
			name:           "incomplete chart - missing templates",
			chartDir:       "incomplete-chart",
			expectedResult: false,
		},
		{
			name:           "missing Chart.yaml",
			chartDir:       "missing-chart-yaml",
			expectedResult: false,
		},
		{
			name:           "missing values.yaml",
			chartDir:       "missing-values-yaml",
			expectedResult: false,
		},
		{
			name:           "templates is a file not directory",
			chartDir:       "templates-is-file",
			expectedResult: false,
		},
		{
			name:           "not a chart directory",
			chartDir:       "not-a-chart",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tempDir, tt.chartDir)
			result := IsHelmChart(testDir)
			if result != tt.expectedResult {
				t.Errorf("IsHelmChart('%s') = %v, expected %v", tt.chartDir, result, tt.expectedResult)
			}
		})
	}
}
