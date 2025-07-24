package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dminnear-rh/patternizer/internal/fileutils"
	"github.com/dminnear-rh/patternizer/internal/helm"
	"github.com/dminnear-rh/patternizer/internal/pattern"
)

// runInit handles the initialization logic for the init command.
func runInit(withSecrets bool) error {
	// Get pattern name and repository root
	patternName, repoRoot, err := pattern.GetPatternNameAndRepoRoot()
	if err != nil {
		return fmt.Errorf("error getting pattern information: %w", err)
	}

	// Find Helm charts in the repository
	chartPaths, err := helm.FindTopLevelCharts(repoRoot)
	if err != nil {
		return fmt.Errorf("error finding Helm charts: %w", err)
	}

	// Process values-global.yaml
	actualPatternName, clusterGroupName, err := pattern.ProcessGlobalValues(patternName, repoRoot, withSecrets)
	if err != nil {
		return fmt.Errorf("error processing global values: %w", err)
	}

	// Process cluster group values using the actual pattern name and cluster group name from the global values
	if err := pattern.ProcessClusterGroupValues(actualPatternName, clusterGroupName, repoRoot, chartPaths, withSecrets); err != nil {
		return fmt.Errorf("error processing cluster group values: %w", err)
	}

	// Copy pattern.sh and Makefile from resources
	resourcesDir, err := fileutils.GetResourcesPath()
	if err != nil {
		return fmt.Errorf("error getting resource path: %w", err)
	}

	patternShSrc := filepath.Join(resourcesDir, "pattern.sh")
	patternShDst := filepath.Join(repoRoot, "pattern.sh")
	if err := fileutils.CopyFile(patternShSrc, patternShDst); err != nil {
		return fmt.Errorf("error copying pattern.sh: %w", err)
	}

	// Always copy Makefile-pattern to the pattern repo
	makefilePatternSrc := filepath.Join(resourcesDir, "Makefile-pattern")
	makefilePatternDst := filepath.Join(repoRoot, "Makefile-pattern")
	if err := fileutils.CopyFile(makefilePatternSrc, makefilePatternDst); err != nil {
		return fmt.Errorf("error copying Makefile-pattern: %w", err)
	}

	// Create a simple Makefile that includes Makefile-pattern (only if it doesn't exist)
	makefileSrc := filepath.Join(resourcesDir, "Makefile")
	makefileDst := filepath.Join(repoRoot, "Makefile")
	if _, err := os.Stat(makefileDst); os.IsNotExist(err) {
		if err := fileutils.CopyFile(makefileSrc, makefileDst); err != nil {
			return fmt.Errorf("error copying Makefile: %w", err)
		}
	}

	// Handle secrets setup if requested
	if withSecrets {
		if err := fileutils.HandleSecretsSetup(resourcesDir, repoRoot); err != nil {
			return fmt.Errorf("error setting up secrets: %w", err)
		}
	}

	fmt.Printf("Successfully initialized pattern '%s' in %s\n", actualPatternName, repoRoot)
	if withSecrets {
		fmt.Println("Secrets configuration has been enabled.")
	}

	return nil
}
