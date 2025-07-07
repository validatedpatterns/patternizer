package fileutils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

// ModifyPatternShScript modifies the pattern.sh script to set USE_SECRETS to the desired value.
func ModifyPatternShScript(patternShPath string, useSecrets bool) error {
	file, err := os.Open(patternShPath)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	// Regex to match the USE_SECRETS line
	regex := regexp.MustCompile(`^\s*:\s*"\$\{USE_SECRETS:=(.+)\}"`)

	for i, line := range lines {
		if matches := regex.FindStringSubmatch(line); matches != nil {
			if useSecrets {
				lines[i] = strings.Replace(line, matches[1], "true", 1)
			} else {
				lines[i] = strings.Replace(line, matches[1], "false", 1)
			}
			break
		}
	}

	output, err := os.Create(patternShPath)
	if err != nil {
		return err
	}
	defer output.Close()

	for _, line := range lines {
		_, err := output.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// ModifyMakefileScript modifies the Makefile to set USE_SECRETS to the desired value.
func ModifyMakefileScript(makefilePath string, useSecrets bool) error {
	file, err := os.Open(makefilePath)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	// Regex to match the USE_SECRETS line in Makefile format
	regex := regexp.MustCompile(`^USE_SECRETS\s*\?=\s*(.+)$`)

	for i, line := range lines {
		if matches := regex.FindStringSubmatch(line); matches != nil {
			if useSecrets {
				lines[i] = "USE_SECRETS ?= true"
			} else {
				lines[i] = "USE_SECRETS ?= false"
			}
			break
		}
	}

	output, err := os.Create(makefilePath)
	if err != nil {
		return err
	}
	defer output.Close()

	for _, line := range lines {
		_, err := output.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// HandleSecretsSetup handles the setup for secrets usage by copying the secrets template
// and modifying the pattern.sh script.
func HandleSecretsSetup(resourcesDir, repoRoot string) (err error) {
	// Copy the values-secret.yaml.template file to the pattern root
	secretsTemplateSrc := filepath.Join(resourcesDir, "values-secret.yaml.template")
	secretsTemplateDst := filepath.Join(repoRoot, "values-secret.yaml.template")

	if err = CopyFile(secretsTemplateSrc, secretsTemplateDst); err != nil {
		return fmt.Errorf("error copying secrets template: %w", err)
	}

	// Modify the pattern.sh script to set USE_SECRETS=true
	patternShPath := filepath.Join(repoRoot, "pattern.sh")
	if err = ModifyPatternShScript(patternShPath, true); err != nil {
		return fmt.Errorf("error modifying pattern.sh for secrets: %w", err)
	}

	return nil
}

// GetResourcePath returns the path to the resources directory.
// It checks the PATTERNIZER_RESOURCES_DIR environment variable first,
// and falls back to the current working directory.
func GetResourcePath() (path string, err error) {
	path = os.Getenv("PATTERNIZER_RESOURCES_DIR")
	if path != "" {
		return path, nil
	}

	// Fall back to current directory
	path, err = os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return path, nil
}
