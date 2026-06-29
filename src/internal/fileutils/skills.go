package fileutils

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/validatedpatterns/patternizer/internal/embedded"
)

var skillTargets = []string{".claude", ".cursor"}

// InstallSkills copies all embedded skill directories into the .claude and .cursor skill directories under the given repository root.
func InstallSkills(repoRoot string) error {
	entries, err := fs.ReadDir(embedded.Skills, "skills")
	if err != nil {
		return fmt.Errorf("error reading embedded skills: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()

		for _, target := range skillTargets {
			skillDst := filepath.Join(repoRoot, target, "skills", skillName)
			if err := WriteEmbeddedDir(embedded.Skills, "skills/"+skillName, skillDst); err != nil {
				return fmt.Errorf("error installing skill %s to %s: %w", skillName, target, err)
			}
		}

		fmt.Printf("Installed skill '%s'\n", skillName)
	}

	return nil
}
