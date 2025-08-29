package fileutils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeFileWithMode(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil { // initial perms don't matter
		t.Fatalf("write file failed: %v", err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}
}

func TestCopyFile_CopiesContentsAndMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	content := "hello world"
	srcMode := os.FileMode(0o640)
	writeFileWithMode(t, src, content, srcMode)

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst failed: %v", err)
	}
	if string(got) != content {
		t.Fatalf("unexpected content: %q", string(got))
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst failed: %v", err)
	}
	// Compare permissions only (mask out non-permission bits)
	if info.Mode().Perm() != srcMode.Perm() {
		t.Fatalf("mode mismatch: got %v want %v", info.Mode().Perm(), srcMode.Perm())
	}
}

func TestHandleSecretsSetup_CopiesWhenMissing_DoesNotOverwriteWhenPresent(t *testing.T) {
	resources := t.TempDir()
	repoRoot := t.TempDir()

	templatePath := filepath.Join(resources, "values-secret.yaml.template")
	originalContent := "foo: bar\n"
	if err := os.WriteFile(templatePath, []byte(originalContent), 0o644); err != nil {
		t.Fatalf("write template failed: %v", err)
	}

	// First call should copy
	if err := HandleSecretsSetup(resources, repoRoot); err != nil {
		t.Fatalf("HandleSecretsSetup failed: %v", err)
	}
	copied := filepath.Join(repoRoot, "values-secret.yaml.template")
	data, err := os.ReadFile(copied)
	if err != nil {
		t.Fatalf("read copied failed: %v", err)
	}
	if string(data) != originalContent {
		t.Fatalf("unexpected copied content: %q", string(data))
	}

	// Change the source and call again; destination should remain unchanged
	if err := os.WriteFile(templatePath, []byte("baz: qux\n"), 0o644); err != nil {
		t.Fatalf("rewrite template failed: %v", err)
	}
	if err := HandleSecretsSetup(resources, repoRoot); err != nil {
		t.Fatalf("HandleSecretsSetup second call failed: %v", err)
	}
	data2, err := os.ReadFile(copied)
	if err != nil {
		t.Fatalf("read copied again failed: %v", err)
	}
	if string(data2) != originalContent {
		t.Fatalf("destination was overwritten unexpectedly: %q", string(data2))
	}
}

func TestGetResourcesPath_EnvSetAndUnset(t *testing.T) {
	old := os.Getenv("PATTERNIZER_RESOURCES_DIR")
	t.Cleanup(func() { _ = os.Setenv("PATTERNIZER_RESOURCES_DIR", old) })

	tmp := t.TempDir()
	if err := os.Setenv("PATTERNIZER_RESOURCES_DIR", tmp); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	got, err := GetResourcesPath()
	if err != nil || got != tmp {
		t.Fatalf("GetResourcesPath with env set failed: got %q err %v", got, err)
	}

	if err := os.Unsetenv("PATTERNIZER_RESOURCES_DIR"); err != nil {
		t.Fatalf("unsetenv failed: %v", err)
	}
	if _, err := GetResourcesPath(); err == nil {
		t.Fatalf("expected error when env is unset")
	}
}

func TestRemovePathIfExists_FileDirSymlink(t *testing.T) {
	base := t.TempDir()

	// File removal
	f := filepath.Join(base, "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	if err := RemovePathIfExists(f); err != nil {
		t.Fatalf("RemovePathIfExists(file) failed: %v", err)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Fatalf("file not removed")
	}

	// Directory removal
	d := filepath.Join(base, "dir")
	if err := os.MkdirAll(filepath.Join(d, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := RemovePathIfExists(d); err != nil {
		t.Fatalf("RemovePathIfExists(dir) failed: %v", err)
	}
	if _, err := os.Stat(d); !os.IsNotExist(err) {
		t.Fatalf("dir not removed")
	}

	// Symlink removal (link to a directory)
	targetDir := t.TempDir()
	link := filepath.Join(base, "link")
	// Windows symlinks require admin/dev mode; skip on windows
	if runtime.GOOS != "windows" {
		if err := os.Symlink(targetDir, link); err != nil {
			t.Fatalf("symlink failed: %v", err)
		}
		if err := RemovePathIfExists(link); err != nil {
			t.Fatalf("RemovePathIfExists(symlink) failed: %v", err)
		}
		if _, err := os.Lstat(link); !os.IsNotExist(err) {
			t.Fatalf("symlink not removed")
		}
		// Ensure target still exists
		if _, err := os.Stat(targetDir); err != nil {
			t.Fatalf("target dir should still exist: %v", err)
		}
	}

	// Non-existent path should be no-op
	if err := RemovePathIfExists(filepath.Join(base, "does-not-exist")); err != nil {
		t.Fatalf("RemovePathIfExists(nonexistent) failed: %v", err)
	}
}

func TestFileContainsIncludeMakefileCommon_Detection(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "Makefile")

	cases := []struct {
		content string
		want    bool
	}{
		{content: "all:\n\t@echo hi\n", want: false},
		{content: "include Makefile-common\nall:\n\t@echo hi\n", want: true},
		{content: "  include   Makefile-common\nall:\n\t@echo hi\n", want: true},
		{content: "# include Makefile-common\nall:\n\t@echo hi\n", want: false},
		{content: "foo:\n\t@echo foo\n# comment\nbar:\n\t@echo bar\n", want: false},
		{content: strings.Join([]string{"foo:", "\t@echo foo", "include Makefile-common", "bar:", "\t@echo bar", ""}, "\n"), want: true},
	}

	for i, tc := range cases {
		if err := os.WriteFile(p, []byte(tc.content), 0o644); err != nil {
			t.Fatalf("write case %d failed: %v", i, err)
		}
		got, err := FileContainsIncludeMakefileCommon(p)
		if err != nil {
			t.Fatalf("case %d: err: %v", i, err)
		}
		if got != tc.want {
			t.Fatalf("case %d: got %v want %v", i, got, tc.want)
		}
	}
}

func TestPrependLineToFile_PrependsAndPreservesMode(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "Makefile")
	original := "all:\n\t@echo hi\n"
	mode := os.FileMode(0o600)
	writeFileWithMode(t, p, original, mode)

	line := "include Makefile-common"
	if err := PrependLineToFile(p, line); err != nil {
		t.Fatalf("PrependLineToFile failed: %v", err)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	expected := line + "\n" + original
	if string(data) != expected {
		t.Fatalf("unexpected content after prepend: %q", string(data))
	}

	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Mode().Perm() != mode.Perm() {
		t.Fatalf("mode not preserved: got %v want %v", info.Mode().Perm(), mode.Perm())
	}
}
