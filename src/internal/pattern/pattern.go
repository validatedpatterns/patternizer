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

// deprecatedClusterGroupNameField is the legacy key under 'main' that has been
// replaced by 'main.variant'.
const deprecatedClusterGroupNameField = "clusterGroupName"

// ProcessGlobalValues processes the global values YAML file.
// It returns the pattern name and variant that should be used (from the file if they exist, or the detected/default names).
// If the existing file uses the deprecated 'main.clusterGroupName' field, it is automatically migrated to 'main.variant'.
func ProcessGlobalValues(patternName, repoRoot string, withSecrets bool) (actualPatternName, variant string, err error) {
	globalValuesPath := filepath.Join(repoRoot, "values-global.yaml")
	values := types.NewDefaultValuesGlobal()

	yamlFile, err := os.ReadFile(globalValuesPath)
	if err != nil && !os.IsNotExist(err) {
		return "", "", fmt.Errorf("failed to read %s: %w", globalValuesPath, err)
	}

	var fileValues types.ValuesGlobal
	if err == nil {
		if err = yaml.Unmarshal(yamlFile, values); err != nil {
			return "", "", fmt.Errorf("failed to unmarshal YAML from %s: %w", globalValuesPath, err)
		}
		if err = yaml.Unmarshal(yamlFile, &fileValues); err != nil {
			return "", "", fmt.Errorf("failed to unmarshal YAML from %s: %w", globalValuesPath, err)
		}
	}

	migrateDeprecatedVariant(values, fileValues.Main.Variant != "")

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

	return values.Global.Pattern, values.Main.Variant, nil
}

// migrateDeprecatedVariant migrates a legacy 'main.clusterGroupName' field to
// 'main.variant'. An explicit 'main.variant' value (variantExplicit) always wins;
// otherwise the deprecated value is adopted so the existing values-<variant>.yaml
// keeps being used. The deprecated key is always removed so it is not written back.
func migrateDeprecatedVariant(values *types.ValuesGlobal, variantExplicit bool) {
	if values.Main.OtherFields == nil {
		return
	}

	legacy, ok := values.Main.OtherFields[deprecatedClusterGroupNameField]
	if !ok {
		return
	}

	delete(values.Main.OtherFields, deprecatedClusterGroupNameField)

	if !variantExplicit {
		if legacyName, ok := legacy.(string); ok {
			values.Main.Variant = legacyName
		}
	}

	fmt.Printf("Migrated deprecated 'main.clusterGroupName' to 'main.variant' (%s) in values-global.yaml.\n", values.Main.Variant)
}

// UsesDeprecatedClusterGroupName reports whether the pattern's values-global.yaml
// still uses the deprecated 'main.clusterGroupName' field. It returns false without
// error when the file does not exist.
func UsesDeprecatedClusterGroupName(repoRoot string) (bool, error) {
	globalValuesPath := filepath.Join(repoRoot, "values-global.yaml")

	yamlFile, err := os.ReadFile(globalValuesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read %s: %w", globalValuesPath, err)
	}

	var values types.ValuesGlobal
	if err = yaml.Unmarshal(yamlFile, &values); err != nil {
		return false, fmt.Errorf("failed to unmarshal YAML from %s: %w", globalValuesPath, err)
	}

	_, ok := values.Main.OtherFields[deprecatedClusterGroupNameField]
	return ok, nil
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
