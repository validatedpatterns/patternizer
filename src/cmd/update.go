package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dminnear-rh/patternizer/internal/fileutils"
	"github.com/dminnear-rh/patternizer/internal/pattern"
)

// runUpdate handles the update logic for the update command.
func runUpdate(noSecrets bool) error {
	// Get pattern name and repository root
	patternName, repoRoot, err := pattern.GetPatternNameAndRepoRoot()
	if err != nil {
		return fmt.Errorf("error getting pattern information: %w", err)
	}

	// Remove the existing pattern.sh file if it exists (it might be a symlink to common/)
	patternShDst := filepath.Join(repoRoot, "pattern.sh")
	if _, err := os.Stat(patternShDst); err == nil {
		fmt.Printf("Removing existing pattern.sh: %s\n", patternShDst)
		if err := os.Remove(patternShDst); err != nil {
			return fmt.Errorf("error removing existing pattern.sh: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking existing pattern.sh: %w", err)
	}

	// Remove the existing Makefile if it exists (utility container provides its own)
	makefilePath := filepath.Join(repoRoot, "Makefile")
	if _, err := os.Stat(makefilePath); err == nil {
		fmt.Printf("Removing existing Makefile: %s\n", makefilePath)
		if err := os.Remove(makefilePath); err != nil {
			return fmt.Errorf("error removing existing Makefile: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking existing Makefile: %w", err)
	}

	// Delete the common/ directory if it exists
	commonDir := filepath.Join(repoRoot, "common")
	if _, err := os.Stat(commonDir); err == nil {
		fmt.Printf("Removing common/ directory: %s\n", commonDir)
		if err := os.RemoveAll(commonDir); err != nil {
			return fmt.Errorf("error removing common/ directory: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking common/ directory: %w", err)
	}

	// Copy pattern.sh from resources
	resourcesDir, err := fileutils.GetResourcePath()
	if err != nil {
		return fmt.Errorf("error getting resource path: %w", err)
	}

	patternShSrc := filepath.Join(resourcesDir, "pattern.sh")
	if err := fileutils.CopyFile(patternShSrc, patternShDst); err != nil {
		return fmt.Errorf("error copying pattern.sh: %w", err)
	}

	// Set USE_SECRETS in pattern.sh based on the flag
	// By default, update uses secrets (opposite of init)
	useSecrets := !noSecrets
	if err := fileutils.ModifyPatternShScript(patternShDst, useSecrets); err != nil {
		return fmt.Errorf("error modifying pattern.sh: %w", err)
	}

	fmt.Printf("Successfully updated pattern '%s' in %s\n", patternName, repoRoot)
	if useSecrets {
		fmt.Println("Secrets configuration is enabled (USE_SECRETS=true).")
	} else {
		fmt.Println("Secrets configuration is disabled (USE_SECRETS=false).")
	}

	return nil
}
