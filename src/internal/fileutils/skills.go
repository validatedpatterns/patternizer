package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
)

var skillTargets = []string{".claude", ".cursor"}

func InstallSkills(repoRoot string) error {
	skillsDir, err := GetSkillsPath()
	if err != nil {
		return fmt.Errorf("error getting skills path: %w", err)
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return fmt.Errorf("error reading skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillSrc := filepath.Join(skillsDir, skillName)

		for _, target := range skillTargets {
			skillDst := filepath.Join(repoRoot, target, "skills", skillName)
			if err := CopyDir(skillSrc, skillDst); err != nil {
				return fmt.Errorf("error installing skill %s to %s: %w", skillName, target, err)
			}
		}

		fmt.Printf("Installed skill '%s'\n", skillName)
	}

	return nil
}
