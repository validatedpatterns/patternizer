package pattern

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/dminnear-rh/patternizer/internal/fileutils"
	"github.com/dminnear-rh/patternizer/internal/types"
)

// GetPatternNameAndRepoRoot returns the pattern name and repository root directory.
// It attempts to detect the pattern name from the Git repository URL,
// falling back to the directory name if Git is not available.
func GetPatternNameAndRepoRoot() (patternName, repoRoot string, err error) {
	// Get the current working directory
	repoRoot, err = os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Use the basename as the pattern name
	patternName = filepath.Base(repoRoot)
	return patternName, repoRoot, nil
}

// ProcessGlobalValues processes the global values YAML file.
// It returns the pattern name and cluster group name that should be used (from the file if they exist, or the detected/default names).
func ProcessGlobalValues(patternName, repoRoot string, withSecrets bool) (actualPatternName, clusterGroupName string, err error) {
	globalValuesPath := filepath.Join(repoRoot, "values-global.yaml")
	values := types.NewDefaultValuesGlobal()

	// Try to read existing file
	yamlFile, err := os.ReadFile(globalValuesPath)
	if err != nil && !os.IsNotExist(err) {
		return "", "", fmt.Errorf("failed to read %s: %w", globalValuesPath, err)
	}

	if err == nil {
		// File exists, unmarshal into our defaults (natural merging)
		if err = yaml.Unmarshal(yamlFile, values); err != nil {
			return "", "", fmt.Errorf("failed to unmarshal YAML from %s: %w", globalValuesPath, err)
		}
	}

	// Set pattern name if not already set
	if values.Global.Pattern == "" {
		values.Global.Pattern = patternName
	}

	// Set secretLoader.disabled based on withSecrets flag
	// If withSecrets is true, we want secretLoader to be enabled (disabled = false)
	// If withSecrets is false, we want secretLoader to be disabled (disabled = true)
	values.Global.SecretLoader.Disabled = !withSecrets

	// Write back the merged values with 2-space indentation
	if err = fileutils.WriteYAMLWithIndent(values, globalValuesPath); err != nil {
		return "", "", fmt.Errorf("failed to write to %s: %w", globalValuesPath, err)
	}

	return values.Global.Pattern, values.Main.ClusterGroupName, nil
}

// ProcessClusterGroupValues processes the cluster group values YAML file.
func ProcessClusterGroupValues(patternName, clusterGroupName, repoRoot string, chartPaths []string, useSecrets bool) error {
	clusterGroupValuesPath := filepath.Join(repoRoot, fmt.Sprintf("values-%s.yaml", clusterGroupName))
	values := types.NewDefaultValuesClusterGroup(patternName, clusterGroupName, chartPaths, useSecrets)

	// Try to read existing file
	yamlFile, err := os.ReadFile(clusterGroupValuesPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", clusterGroupValuesPath, err)
	}

	if err == nil {
		// File exists, unmarshal into a separate struct first
		var existingValues types.ValuesClusterGroup
		if err = yaml.Unmarshal(yamlFile, &existingValues); err != nil {
			return fmt.Errorf("failed to unmarshal YAML from %s: %w", clusterGroupValuesPath, err)
		}

		// Merge existing values with new defaults intelligently
		mergeClusterGroupValues(values, &existingValues)
	}

	// Write back the merged values with 2-space indentation
	if err = fileutils.WriteYAMLWithIndent(values, clusterGroupValuesPath); err != nil {
		return fmt.Errorf("failed to write to %s: %w", clusterGroupValuesPath, err)
	}

	return nil
}

// mergeClusterGroupValues intelligently merges existing values with new defaults
func mergeClusterGroupValues(defaults, existing *types.ValuesClusterGroup) {
	// Preserve existing applications and merge with new ones
	for key, app := range existing.ClusterGroup.Applications {
		defaults.ClusterGroup.Applications[key] = app
	}

	// For namespaces: preserve existing ones and add secrets-related ones if needed
	existingNamespaceMap := make(map[string]bool)
	for _, ns := range existing.ClusterGroup.Namespaces {
		// Add existing namespace to defaults if not already present
		found := false
		for _, defaultNs := range defaults.ClusterGroup.Namespaces {
			if ns.Equal(defaultNs) {
				found = true
				break
			}
		}
		if !found {
			defaults.ClusterGroup.Namespaces = append(defaults.ClusterGroup.Namespaces, ns)
		}
		// Track what we have
		if nsStr, ok := ns.GetString(); ok {
			existingNamespaceMap[nsStr] = true
		}
	}

	// For projects: preserve existing ones and add cluster group project if secrets are needed
	existingProjectMap := make(map[string]bool)
	for _, proj := range existing.ClusterGroup.Projects {
		existingProjectMap[proj] = true
	}

	// Rebuild projects list preserving existing order but ensuring required projects are present
	mergedProjects := make([]string, 0)

	// Add existing projects first
	mergedProjects = append(mergedProjects, existing.ClusterGroup.Projects...)

	// Add any missing required projects
	for _, proj := range defaults.ClusterGroup.Projects {
		if !existingProjectMap[proj] {
			mergedProjects = append(mergedProjects, proj)
		}
	}

	defaults.ClusterGroup.Projects = mergedProjects

	// Preserve other fields from existing
	if existing.ClusterGroup.IsHubCluster {
		defaults.ClusterGroup.IsHubCluster = existing.ClusterGroup.IsHubCluster
	}

	// Merge subscriptions
	for key, sub := range existing.ClusterGroup.Subscriptions {
		defaults.ClusterGroup.Subscriptions[key] = sub
	}

	// Merge other fields
	if existing.ClusterGroup.OtherFields != nil {
		for key, value := range existing.ClusterGroup.OtherFields {
			defaults.ClusterGroup.OtherFields[key] = value
		}
	}
	if existing.OtherFields != nil {
		for key, value := range existing.OtherFields {
			defaults.OtherFields[key] = value
		}
	}
}
