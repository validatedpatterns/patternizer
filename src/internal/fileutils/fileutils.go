package fileutils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CopyFile copies a file from src to dst. If dst already exists, it will be overwritten.
// The function also preserves the file permissions of the source file.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source file %s: %w", src, err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file %s: %w", dst, err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("copy to %s: %w", dst, err)
	}

	err = os.Chmod(dst, sourceFileStat.Mode())
	if err != nil {
		return fmt.Errorf("chmod %s: %w", dst, err)
	}

	return nil
}

// HandleSecretsSetup handles the setup for secrets usage by copying the secrets template.
func HandleSecretsSetup(fsys fs.FS, repoRoot string) error {
	secretsTemplateDst := filepath.Join(repoRoot, "values-secret.yaml.template")

	if _, err := os.Stat(secretsTemplateDst); os.IsNotExist(err) {
		if err = WriteEmbeddedFile(fsys, "resources/values-secret.yaml.template", secretsTemplateDst, 0o644); err != nil {
			return fmt.Errorf("error copying secrets template: %w", err)
		}
	}

	return nil
}

// WriteEmbeddedFile reads a file from an embedded FS and writes it to disk with the given mode.
func WriteEmbeddedFile(fsys fs.FS, srcPath, dstPath string, mode os.FileMode) error {
	data, err := fs.ReadFile(fsys, srcPath)
	if err != nil {
		return fmt.Errorf("reading embedded file %s: %w", srcPath, err)
	}
	if err := os.WriteFile(dstPath, data, mode); err != nil {
		return fmt.Errorf("write file %s: %w", dstPath, err)
	}
	return os.Chmod(dstPath, mode)
}

// WriteEmbeddedDir recursively copies an embedded directory tree to disk.
func WriteEmbeddedDir(fsys fs.FS, srcDir, dstDir string) error {
	return fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk embedded dir: %w", err)
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("compute relative path for %s: %w", path, err)
		}
		target := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return WriteEmbeddedFile(fsys, path, target, 0o644)
	})
}

// CopyDir recursively copies the contents of src into dst.
// It creates dst and any necessary subdirectories.
// Existing files in dst are overwritten, but files not present in src are left untouched.
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source dir %s: %w", src, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create destination dir %s: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read source dir %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("copy subdir %s: %w", entry.Name(), err)
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("copy file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
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
		return fmt.Errorf("lstat %s: %w", targetPath, err)
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
		return false, fmt.Errorf("read %s: %w", makefilePath, err)
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
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(filePath); statErr == nil {
		mode = info.Mode()
	}

	newContents := []byte(line + "\n" + string(data))
	return os.WriteFile(filePath, newContents, mode)
}

// WriteYAMLWithIndent marshals the given data structure to YAML and writes it to a file
// with 2-space indentation. This ensures consistency with prettier formatting.
func WriteYAMLWithIndent(data interface{}, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	// Set indentation to 2 spaces instead of the default 4
	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML to %s: %w", filePath, err)
	}

	if err := os.Chmod(filePath, 0o644); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", filePath, err)
	}

	return nil
}
