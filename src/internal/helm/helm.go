package helm

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// IsHelmChart checks if a given directory path contains a valid Helm chart.
func IsHelmChart(path string) bool {
	chartYamlPath := filepath.Join(path, "Chart.yaml")
	valuesYamlPath := filepath.Join(path, "values.yaml")
	templatesDirPath := filepath.Join(path, "templates")

	_, chartErr := os.Stat(chartYamlPath)
	_, valuesErr := os.Stat(valuesYamlPath)
	templatesInfo, templatesErr := os.Stat(templatesDirPath)

	// If any of the essential files don't exist, it's not a chart.
	if os.IsNotExist(chartErr) || os.IsNotExist(valuesErr) || os.IsNotExist(templatesErr) {
		return false
	}
	// The 'templates' path must be a directory.
	if !templatesInfo.IsDir() {
		return false
	}
	return true
}

// FindTopLevelCharts walks the filesystem from rootDir to find all top-level Helm charts.
// It intelligently skips sub-chart directories.
func FindTopLevelCharts(rootDir string) ([]string, error) {
	var charts []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		// Skip hidden directories (like .git) and the charts directory itself to avoid recursion
		if strings.HasPrefix(d.Name(), ".") || (d.Name() == "charts" && path != filepath.Join(rootDir, "charts")) {
			return filepath.SkipDir
		}

		if IsHelmChart(path) {
			relPath, _ := filepath.Rel(rootDir, path)
			charts = append(charts, relPath)
			// Once we identify a chart, we don't need to look at its subdirectories.
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return charts, nil
}
