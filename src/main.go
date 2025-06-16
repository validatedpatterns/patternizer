package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// getPatternNameAndRepoRoot determines the pattern's canonical name from the git remote URL.
// It also returns the local path to the repository root for filesystem scans.
func getPatternNameAndRepoRoot() (string, string, error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return "", "", fmt.Errorf("could not find repo root: %w", err)
	}

	urlBytes, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		log.Printf("Could not get git remote 'origin'. Falling back to local directory name.")
		patternName := filepath.Base(repoRoot)
		return patternName, repoRoot, nil
	}

	urlString := strings.TrimSpace(string(urlBytes))
	nameWithSuffix := filepath.Base(urlString)
	patternName := strings.TrimSuffix(nameWithSuffix, ".git")

	return patternName, repoRoot, nil
}

// getRepoRoot finds the top-level directory of the current git repository.
func getRepoRoot() (string, error) {
	pathBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("not a git repository: %s", string(exitError.Stderr))
		}
		return "", fmt.Errorf("could not execute git command: %w", err)
	}
	return strings.TrimSpace(string(pathBytes)), nil
}

// processGlobalValues handles the creation and updating of the values-global.yaml file.
func processGlobalValues(patternName string) (*ValuesGlobal, error) {
	const filename = "values-global.yaml"
	values := newDefaultValuesGlobal()

	yamlFile, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	if err == nil {
		log.Printf("Found existing '%s', reading and merging.", filename)
		if err = yaml.Unmarshal(yamlFile, values); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML from %s: %w", filename, err)
		}
	} else {
		log.Printf("'%s' not found, will create with default values.", filename)
	}

	if values.Global.Pattern == "" {
		values.Global.Pattern = patternName
	}

	finalYamlBytes, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal global values: %w", err)
	}
	if err = os.WriteFile(filename, finalYamlBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write to %s: %w", filename, err)
	}

	log.Printf("Successfully processed '%s'.", filename)
	return values, nil
}

// processClusterGroupValues handles the creation/update of the values-<clustergroup>.yaml file.
func processClusterGroupValues(globalValues *ValuesGlobal, repoRoot string) error {
	clusterGroupName := globalValues.Main.ClusterGroupName
	patternName := globalValues.Global.Pattern
	filename := fmt.Sprintf("values-%s.yaml", clusterGroupName)

	chartPaths, err := findTopLevelCharts(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to find helm charts: %w", err)
	}
	log.Printf("Found %d top-level charts.", len(chartPaths))

	values := newDefaultValuesClusterGroup(patternName, clusterGroupName, chartPaths)

	yamlFile, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", filename, err)
	}

	if err == nil {
		log.Printf("Found existing '%s', reading and merging.", filename)
		if err = yaml.Unmarshal(yamlFile, values); err != nil {
			return fmt.Errorf("failed to unmarshal YAML from %s: %w", filename, err)
		}
	} else {
		log.Printf("'%s' not found, will create with default values.", filename)
	}

	finalYamlBytes, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal cluster group values: %w", err)
	}
	if err = os.WriteFile(filename, finalYamlBytes, 0644); err != nil {
		return fmt.Errorf("failed to write to %s: %w", filename, err)
	}

	log.Printf("Successfully processed '%s'.", filename)
	return nil
}

func main() {
	patternName, repoRoot, err := getPatternNameAndRepoRoot()
	if err != nil {
		log.Fatalf("Error determining pattern name or repo root: %v", err)
	}

	log.Printf("Determined pattern name: '%s'", patternName)

	globalValues, err := processGlobalValues(patternName)
	if err != nil {
		log.Fatalf("Error processing global values: %v", err)
	}

	if err := processClusterGroupValues(globalValues, repoRoot); err != nil {
		log.Fatalf("Error processing cluster group values: %v", err)
	}

	log.Println("All configuration files processed successfully.")
}
