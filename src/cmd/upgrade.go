package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/validatedpatterns/patternizer/internal/embedded"
	"github.com/validatedpatterns/patternizer/internal/fileutils"
	"github.com/validatedpatterns/patternizer/internal/pattern"
)

// runUpgrade handles the upgrade logic for the upgrade command.
func runUpgrade(replaceMakefile bool) error {
	_, repoRoot, err := pattern.GetPatternNameAndRepoRoot()
	if err != nil {
		return fmt.Errorf("error getting pattern information: %w", err)
	}

	commonDirPath := filepath.Join(repoRoot, "common")
	patternShPath := filepath.Join(repoRoot, "pattern.sh")

	if err := fileutils.RemovePathIfExists(commonDirPath); err != nil {
		return fmt.Errorf("error removing common directory: %w", err)
	}

	if err := fileutils.RemovePathIfExists(patternShPath); err != nil {
		return fmt.Errorf("error removing pattern.sh: %w", err)
	}

	if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/pattern.sh", patternShPath, 0o755); err != nil {
		return fmt.Errorf("error copying pattern.sh: %w", err)
	}

	if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/Makefile-common", filepath.Join(repoRoot, "Makefile-common"), 0o644); err != nil {
		return fmt.Errorf("error copying Makefile-common: %w", err)
	}

	if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/ansible.cfg", filepath.Join(repoRoot, "ansible.cfg"), 0o644); err != nil {
		return fmt.Errorf("error copying ansible.cfg: %w", err)
	}

	makefileDst := filepath.Join(repoRoot, "Makefile")

	if replaceMakefile {
		if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/Makefile", makefileDst, 0o644); err != nil {
			return fmt.Errorf("error replacing Makefile: %w", err)
		}
	} else {
		if _, err := os.Stat(makefileDst); os.IsNotExist(err) {
			if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/Makefile", makefileDst, 0o644); err != nil {
				return fmt.Errorf("error copying Makefile: %w", err)
			}
		} else if err == nil {
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

	if err := fileutils.InstallSkills(repoRoot); err != nil {
		return fmt.Errorf("error installing skills: %w", err)
	}

	fmt.Printf("Successfully upgraded pattern repository in %s\n", repoRoot)
	return nil
}
