package pattern

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/validatedpatterns/patternizer/internal/fileutils"
	"github.com/validatedpatterns/patternizer/internal/types"
)

// GetPatternNameAndRepoRoot returns the pattern name and repository root directory.
// The pattern name is derived from the basename of the current working directory.
func GetPatternNameAndRepoRoot() (patternName, repoRoot string, err error) {
	repoRoot, err = os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	patternName = filepath.Base(repoRoot)
	return patternName, repoRoot, nil
}

// ProcessGlobalValues processes the global values YAML file.
// It returns the pattern name and cluster group name that should be used (from the file if they exist, or the detected/default names).
func ProcessGlobalValues(patternName, repoRoot string, withSecrets bool) (actualPatternName, clusterGroupName string, err error) {
	globalValuesPath := filepath.Join(repoRoot, "values-global.yaml")
	values := types.NewDefaultValuesGlobal()

	yamlFile, err := os.ReadFile(globalValuesPath)
	if err != nil && !os.IsNotExist(err) {
		return "", "", fmt.Errorf("failed to read %s: %w", globalValuesPath, err)
	}

	if err == nil {
		if err = yaml.Unmarshal(yamlFile, values); err != nil {
			return "", "", fmt.Errorf("failed to unmarshal YAML from %s: %w", globalValuesPath, err)
		}
	}

	if values.Global.Pattern == "" {
		values.Global.Pattern = patternName
	}

	// Set secretLoader.disabled based on withSecrets flag
	// If withSecrets is true, we want secretLoader to be enabled (disabled = false)
	// If withSecrets is false, we want secretLoader to be disabled (disabled = true)
	values.Global.SecretLoader.Disabled = !withSecrets

	if err = fileutils.WriteYAMLWithIndent(values, globalValuesPath); err != nil {
		return "", "", fmt.Errorf("failed to write to %s: %w", globalValuesPath, err)
	}

	return values.Global.Pattern, values.Main.ClusterGroupName, nil
}

// ProcessClusterGroupValues processes the cluster group values YAML file.
func ProcessClusterGroupValues(patternName, clusterGroupName, repoRoot string, chartPaths []string, useSecrets bool) error {
	clusterGroupValuesPath := filepath.Join(repoRoot, fmt.Sprintf("values-%s.yaml", clusterGroupName))
	values := types.NewDefaultValuesClusterGroup(patternName, clusterGroupName, chartPaths, useSecrets)

	yamlFile, err := os.ReadFile(clusterGroupValuesPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", clusterGroupValuesPath, err)
	}

	if err == nil {
		var existingValues types.ValuesClusterGroup
		if err = yaml.Unmarshal(yamlFile, &existingValues); err != nil {
			return fmt.Errorf("failed to unmarshal YAML from %s: %w", clusterGroupValuesPath, err)
		}

		mergeClusterGroupValues(values, &existingValues)
	}

	if err = fileutils.WriteYAMLWithIndent(values, clusterGroupValuesPath); err != nil {
		return fmt.Errorf("failed to write to %s: %w", clusterGroupValuesPath, err)
	}

	return nil
}

// mergeClusterGroupValues intelligently merges existing values with new defaults
func mergeClusterGroupValues(defaults, existing *types.ValuesClusterGroup) {
	for key, app := range existing.ClusterGroup.Applications {
		defaults.ClusterGroup.Applications[key] = app
	}

	for nsName, nsConfig := range existing.ClusterGroup.Namespaces {
		defaults.ClusterGroup.Namespaces[nsName] = nsConfig
	}

	existingProjectMap := make(map[string]bool)
	for _, proj := range existing.ClusterGroup.Projects {
		existingProjectMap[proj] = true
	}

	mergedProjects := make([]string, 0)
	mergedProjects = append(mergedProjects, existing.ClusterGroup.Projects...)

	for _, proj := range defaults.ClusterGroup.Projects {
		if !existingProjectMap[proj] {
			mergedProjects = append(mergedProjects, proj)
		}
	}

	defaults.ClusterGroup.Projects = mergedProjects

	for key, sub := range existing.ClusterGroup.Subscriptions {
		defaults.ClusterGroup.Subscriptions[key] = sub
	}

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
