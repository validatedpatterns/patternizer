package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// getResourcePath returns the path to a resource file, checking the PATTERNIZER_RESOURCES_DIR
// environment variable first, then falling back to the current directory
func getResourcePath(filename string) string {
	if resourcesDir := os.Getenv("PATTERNIZER_RESOURCES_DIR"); resourcesDir != "" {
		return filepath.Join(resourcesDir, filename)
	}
	return filename
}

// copyFile copies a file from src to dst, preserving file permissions
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	// Get source file info to preserve permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info %s: %w", src, err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Preserve the original file permissions
	err = os.Chmod(dst, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", dst, err)
	}

	return nil
}

// copyPatternScript copies the pattern.sh script to the current directory
func copyPatternScript() error {
	patternShPath := getResourcePath("pattern.sh")

	if _, err := os.Stat(patternShPath); err == nil {
		log.Println("Copying pattern.sh script")
		if err := copyFile(patternShPath, "pattern.sh"); err != nil {
			return fmt.Errorf("failed to copy pattern.sh: %w", err)
		}
	} else {
		return fmt.Errorf("pattern.sh not found at %s", patternShPath)
	}

	return nil
}

// modifyPatternShScript modifies the pattern.sh file to set USE_SECRETS=true by default
func modifyPatternShScript(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read pattern.sh: %w", err)
	}

	modifiedContent := strings.Replace(string(content), "${USE_SECRETS:=false}", "${USE_SECRETS:=true}", 1)

	err = os.WriteFile(filePath, []byte(modifiedContent), 0o755)
	if err != nil {
		return fmt.Errorf("failed to write modified pattern.sh: %w", err)
	}

	return nil
}

// handleSecretsSetup handles copying the secrets template and modifying pattern.sh when --with-secrets is used
func handleSecretsSetup() error {
	templatePath := getResourcePath("values-secret.yaml.template")

	if _, err := os.Stat(templatePath); err == nil {
		log.Println("Copying secrets template")
		if err := copyFile(templatePath, "values-secret.yaml.template"); err != nil {
			return fmt.Errorf("failed to copy secrets template: %w", err)
		}
	} else {
		return fmt.Errorf("secrets template not found at %s", templatePath)
	}

	log.Println("Modifying pattern.sh for secrets usage")
	if err := modifyPatternShScript("pattern.sh"); err != nil {
		return fmt.Errorf("failed to modify pattern.sh: %w", err)
	}

	return nil
}
