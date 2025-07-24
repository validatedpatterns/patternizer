package fileutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	// Fall back to current directory
	path, err = filepath.Abs("resources")
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return path, nil
}
