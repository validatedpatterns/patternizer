package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dminnear-rh/patternizer/internal/fileutils"
	"github.com/dminnear-rh/patternizer/internal/pattern"
)

// runUpgrade handles the upgrade logic for the upgrade command.
func runUpgrade(replaceMakefile bool) error {
	// Determine repository root
	_, repoRoot, err := pattern.GetPatternNameAndRepoRoot()
	if err != nil {
		return fmt.Errorf("error getting pattern information: %w", err)
	}

	// Paths
	commonDirPath := filepath.Join(repoRoot, "common")
	patternShPath := filepath.Join(repoRoot, "pattern.sh")

	// Remove common/ directory if it exists
	if err := fileutils.RemovePathIfExists(commonDirPath); err != nil {
		return fmt.Errorf("error removing common directory: %w", err)
	}

	// Remove ./pattern.sh if it exists (symlink or file)
	if err := fileutils.RemovePathIfExists(patternShPath); err != nil {
		return fmt.Errorf("error removing pattern.sh: %w", err)
	}

	// Copy resources into repo root
	resourcesDir, err := fileutils.GetResourcesPath()
	if err != nil {
		return fmt.Errorf("error getting resource path: %w", err)
	}

	// Copy pattern.sh
	if err := fileutils.CopyFile(filepath.Join(resourcesDir, "pattern.sh"), patternShPath); err != nil {
		return fmt.Errorf("error copying pattern.sh: %w", err)
	}

	// Copy Makefile-common
	if err := fileutils.CopyFile(filepath.Join(resourcesDir, "Makefile-common"), filepath.Join(repoRoot, "Makefile-common")); err != nil {
		return fmt.Errorf("error copying Makefile-common: %w", err)
	}

	// Makefile handling
	makefileSrc := filepath.Join(resourcesDir, "Makefile")
	makefileDst := filepath.Join(repoRoot, "Makefile")

	if replaceMakefile {
		if err := fileutils.CopyFile(makefileSrc, makefileDst); err != nil {
			return fmt.Errorf("error replacing Makefile: %w", err)
		}
	} else {
		// If Makefile doesn't exist, copy it
		if _, err := os.Stat(makefileDst); os.IsNotExist(err) {
			if err := fileutils.CopyFile(makefileSrc, makefileDst); err != nil {
				return fmt.Errorf("error copying Makefile: %w", err)
			}
		} else if err == nil {
			// If Makefile exists, check for include and prepend if missing
			hasInclude, err := fileutils.FileContainsIncludeMakefileCommon(makefileDst)
			if err != nil {
				return fmt.Errorf("error checking Makefile for include: %w", err)
			}
			if !hasInclude {
				if err := fileutils.PrependLineToFile(makefileDst, "include Makefile-common"); err != nil {
					return fmt.Errorf("error updating Makefile: %w", err)
				}
			}
		} else {
			return fmt.Errorf("error accessing Makefile: %w", err)
		}
	}

	fmt.Printf("Successfully upgraded pattern repository in %s\n", repoRoot)
	return nil
}
