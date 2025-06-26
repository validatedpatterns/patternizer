package main

import (
	"fmt"
	"log"
)

// runInit executes the initialization logic
func runInit(withSecrets bool) error {
	patternName, repoRoot, err := getPatternNameAndRepoRoot()
	if err != nil {
		return fmt.Errorf("error determining pattern name or repo root: %w", err)
	}

	log.Printf("Determined pattern name: '%s'", patternName)

	globalValues, err := processGlobalValues(patternName)
	if err != nil {
		return fmt.Errorf("error processing global values: %w", err)
	}

	log.Printf("Secrets will%s be added to the cluster group values file", map[bool]string{true: "", false: " not"}[withSecrets])

	if err := processClusterGroupValues(globalValues, repoRoot, withSecrets); err != nil {
		return fmt.Errorf("error processing cluster group values: %w", err)
	}

	// Always copy pattern.sh
	if err := copyPatternScript(); err != nil {
		return fmt.Errorf("error copying pattern.sh: %w", err)
	}

	// Handle secrets setup if requested
	if withSecrets {
		if err := handleSecretsSetup(); err != nil {
			return fmt.Errorf("error setting up secrets: %w", err)
		}
	}

	log.Println("All configuration files processed successfully.")
	return nil
}
