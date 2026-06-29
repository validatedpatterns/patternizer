package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/validatedpatterns/patternizer/internal/embedded"
	"github.com/validatedpatterns/patternizer/internal/fileutils"
	"github.com/validatedpatterns/patternizer/internal/helm"
	"github.com/validatedpatterns/patternizer/internal/pattern"
)

// runInit handles the initialization logic for the init command.
func runInit(withSecrets bool) error {
	patternName, repoRoot, err := pattern.GetPatternNameAndRepoRoot()
	if err != nil {
		return fmt.Errorf("error getting pattern information: %w", err)
	}

	chartPaths, err := helm.FindTopLevelCharts(repoRoot)
	if err != nil {
		return fmt.Errorf("error finding Helm charts: %w", err)
	}

	actualPatternName, clusterGroupName, err := pattern.ProcessGlobalValues(patternName, repoRoot, withSecrets)
	if err != nil {
		return fmt.Errorf("error processing global values: %w", err)
	}

	if err := pattern.ProcessClusterGroupValues(actualPatternName, clusterGroupName, repoRoot, chartPaths, withSecrets); err != nil {
		return fmt.Errorf("error processing cluster group values: %w", err)
	}

	if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/pattern.sh", filepath.Join(repoRoot, "pattern.sh"), 0o755); err != nil {
		return fmt.Errorf("error copying pattern.sh: %w", err)
	}

	if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/ansible.cfg", filepath.Join(repoRoot, "ansible.cfg"), 0o644); err != nil {
		return fmt.Errorf("error copying ansible.cfg: %w", err)
	}

	if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/Makefile-common", filepath.Join(repoRoot, "Makefile-common"), 0o644); err != nil {
		return fmt.Errorf("error copying Makefile-common: %w", err)
	}

	makefileDst := filepath.Join(repoRoot, "Makefile")
	if _, err := os.Stat(makefileDst); os.IsNotExist(err) {
		if err := fileutils.WriteEmbeddedFile(embedded.Resources, "resources/Makefile", makefileDst, 0o644); err != nil {
			return fmt.Errorf("error copying Makefile: %w", err)
		}
	}

	if withSecrets {
		if err := fileutils.HandleSecretsSetup(embedded.Resources, repoRoot); err != nil {
			return fmt.Errorf("error setting up secrets: %w", err)
		}
	}

	if err := fileutils.InstallSkills(repoRoot); err != nil {
		return fmt.Errorf("error installing skills: %w", err)
	}

	fmt.Printf("Successfully initialized pattern '%s' in %s\n", actualPatternName, repoRoot)
	if withSecrets {
		fmt.Println("Secrets configuration has been enabled.")
	}

	return nil
}
