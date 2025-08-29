package fileutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyFile copies a file from src to dst. If dst already exists, it will be overwritten.
// The function also preserves the file permissions of the source file.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	// Preserve the file permissions from the source file
	err = os.Chmod(dst, sourceFileStat.Mode())
	if err != nil {
		return err
	}

	return nil
}

// HandleSecretsSetup handles the setup for secrets usage by copying the secrets template.
func HandleSecretsSetup(resourcesDir, repoRoot string) (err error) {
	// Copy the values-secret.yaml.template file to the pattern root only if it doesn't already exist
	secretsTemplateSrc := filepath.Join(resourcesDir, "values-secret.yaml.template")
	secretsTemplateDst := filepath.Join(repoRoot, "values-secret.yaml.template")

	if _, err := os.Stat(secretsTemplateDst); os.IsNotExist(err) {
		if err = CopyFile(secretsTemplateSrc, secretsTemplateDst); err != nil {
			return fmt.Errorf("error copying secrets template: %w", err)
		}
	}

	return nil
}

// GetResourcesPath returns the path to the resources directory.
// It checks the PATTERNIZER_RESOURCES_DIR environment variable first,
// and falls back to the current working directory.
func GetResourcesPath() (path string, err error) {
	path = os.Getenv("PATTERNIZER_RESOURCES_DIR")
	if path != "" {
		return path, nil
	}

	// Error out if the resources directory is not found
	return "", fmt.Errorf("PATTERNIZER_RESOURCES_DIR environment variable is not set")
}

// RemovePathIfExists removes a file, directory, or symlink at the given path if it exists.
// It does nothing if the path does not exist.
func RemovePathIfExists(targetPath string) error {
	if targetPath == "" {
		return nil
	}
	info, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if info.IsDir() {
		return os.RemoveAll(targetPath)
	}

	return os.Remove(targetPath)
}

// FileContainsIncludeMakefileCommon checks if a Makefile already contains an include Makefile-common line.
func FileContainsIncludeMakefileCommon(makefilePath string) (bool, error) {
	data, err := os.ReadFile(makefilePath)
	if err != nil {
		return false, err
	}
	contents := string(data)
	// We keep this simple to avoid regex: look for lines with 'include' and 'Makefile-common'
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, "include") && strings.Contains(trimmed, "Makefile-common") {
			return true, nil
		}
	}
	return false, nil
}

// PrependLineToFile prepends a line to a file, preserving existing permissions.
func PrependLineToFile(filePath, line string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(filePath); statErr == nil {
		mode = info.Mode()
	}

	newContents := []byte(line + "\n" + string(data))
	return os.WriteFile(filePath, newContents, mode)
}
