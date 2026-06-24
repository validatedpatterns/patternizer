package fileutils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

func writeFileWithMode(path, content string, mode os.FileMode) {
	Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
	Expect(os.Chmod(path, mode)).To(Succeed())
}

var _ = Describe("CopyFile", func() {
	It("should copy contents and preserve file mode", func() {
		dir := GinkgoT().TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")

		content := "hello world"
		srcMode := os.FileMode(0o640)
		writeFileWithMode(src, content, srcMode)

		Expect(CopyFile(src, dst)).To(Succeed())

		got, err := os.ReadFile(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got)).To(Equal(content))

		info, err := os.Stat(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(srcMode.Perm()))
	})
})

var _ = Describe("HandleSecretsSetup", func() {
	It("should copy template when missing and not overwrite when present", func() {
		resources := GinkgoT().TempDir()
		repoRoot := GinkgoT().TempDir()

		templatePath := filepath.Join(resources, "values-secret.yaml.template")
		originalContent := "foo: bar\n"
		Expect(os.WriteFile(templatePath, []byte(originalContent), 0o644)).To(Succeed())

		// First call should copy
		Expect(HandleSecretsSetup(resources, repoRoot)).To(Succeed())
		copied := filepath.Join(repoRoot, "values-secret.yaml.template")
		data, err := os.ReadFile(copied)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal(originalContent))

		// Change the source and call again; destination should remain unchanged
		Expect(os.WriteFile(templatePath, []byte("baz: qux\n"), 0o644)).To(Succeed())
		Expect(HandleSecretsSetup(resources, repoRoot)).To(Succeed())
		data2, err := os.ReadFile(copied)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data2)).To(Equal(originalContent))
	})
})

var _ = Describe("GetResourcesPath", func() {
	It("should return the path when the environment variable is set", func() {
		old := os.Getenv("PATTERNIZER_RESOURCES_DIR")
		DeferCleanup(func() { os.Setenv("PATTERNIZER_RESOURCES_DIR", old) })

		tmp := GinkgoT().TempDir()
		Expect(os.Setenv("PATTERNIZER_RESOURCES_DIR", tmp)).To(Succeed())
		got, err := GetResourcesPath()
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(tmp))
	})

	It("should return an error when the environment variable is unset", func() {
		old := os.Getenv("PATTERNIZER_RESOURCES_DIR")
		DeferCleanup(func() { os.Setenv("PATTERNIZER_RESOURCES_DIR", old) })

		Expect(os.Unsetenv("PATTERNIZER_RESOURCES_DIR")).To(Succeed())
		_, err := GetResourcesPath()
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("GetSkillsPath", func() {
	It("should return the path when the environment variable is set", func() {
		old := os.Getenv("PATTERNIZER_SKILLS_DIR")
		DeferCleanup(func() { os.Setenv("PATTERNIZER_SKILLS_DIR", old) })

		tmp := GinkgoT().TempDir()
		Expect(os.Setenv("PATTERNIZER_SKILLS_DIR", tmp)).To(Succeed())
		got, err := GetSkillsPath()
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(tmp))
	})

	It("should return an error when the environment variable is unset", func() {
		old := os.Getenv("PATTERNIZER_SKILLS_DIR")
		DeferCleanup(func() { os.Setenv("PATTERNIZER_SKILLS_DIR", old) })

		Expect(os.Unsetenv("PATTERNIZER_SKILLS_DIR")).To(Succeed())
		_, err := GetSkillsPath()
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("CopyDir", func() {
	It("should recursively copy a directory and its contents", func() {
		src := GinkgoT().TempDir()
		dst := filepath.Join(GinkgoT().TempDir(), "dest")

		Expect(os.MkdirAll(filepath.Join(src, "sub"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(src, "file1.txt"), []byte("hello"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(src, "sub", "file2.txt"), []byte("world"), 0o644)).To(Succeed())

		Expect(CopyDir(src, dst)).To(Succeed())

		got1, err := os.ReadFile(filepath.Join(dst, "file1.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got1)).To(Equal("hello"))

		got2, err := os.ReadFile(filepath.Join(dst, "sub", "file2.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got2)).To(Equal("world"))
	})

	It("should overwrite existing files without deleting unrelated files", func() {
		src := GinkgoT().TempDir()
		dst := GinkgoT().TempDir()

		Expect(os.WriteFile(filepath.Join(src, "file.txt"), []byte("new"), 0o644)).To(Succeed())

		Expect(os.WriteFile(filepath.Join(dst, "file.txt"), []byte("old"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dst, "unrelated.txt"), []byte("keep me"), 0o644)).To(Succeed())

		Expect(CopyDir(src, dst)).To(Succeed())

		got, err := os.ReadFile(filepath.Join(dst, "file.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got)).To(Equal("new"))

		kept, err := os.ReadFile(filepath.Join(dst, "unrelated.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(kept)).To(Equal("keep me"))
	})

	It("should return an error when source is not a directory", func() {
		dir := GinkgoT().TempDir()
		file := filepath.Join(dir, "file.txt")
		Expect(os.WriteFile(file, []byte("x"), 0o644)).To(Succeed())

		err := CopyDir(file, filepath.Join(dir, "dst"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("is not a directory"))
	})
})

var _ = Describe("RemovePathIfExists", func() {
	It("should remove a file", func() {
		base := GinkgoT().TempDir()
		f := filepath.Join(base, "file.txt")
		Expect(os.WriteFile(f, []byte("x"), 0o644)).To(Succeed())

		Expect(RemovePathIfExists(f)).To(Succeed())
		_, err := os.Stat(f)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("should remove a directory", func() {
		base := GinkgoT().TempDir()
		d := filepath.Join(base, "dir")
		Expect(os.MkdirAll(filepath.Join(d, "nested"), 0o755)).To(Succeed())

		Expect(RemovePathIfExists(d)).To(Succeed())
		_, err := os.Stat(d)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("should remove a symlink without removing the target", func() {
		if runtime.GOOS == "windows" {
			Skip("symlink tests require admin/dev mode on Windows")
		}

		base := GinkgoT().TempDir()
		targetDir := GinkgoT().TempDir()
		link := filepath.Join(base, "link")
		Expect(os.Symlink(targetDir, link)).To(Succeed())

		Expect(RemovePathIfExists(link)).To(Succeed())
		_, err := os.Lstat(link)
		Expect(os.IsNotExist(err)).To(BeTrue())

		// Ensure target still exists
		_, err = os.Stat(targetDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should be a no-op for a non-existent path", func() {
		base := GinkgoT().TempDir()
		Expect(RemovePathIfExists(filepath.Join(base, "does-not-exist"))).To(Succeed())
	})
})

var _ = Describe("FileContainsIncludeMakefileCommon", func() {
	DescribeTable("should detect include Makefile-common correctly",
		func(content string, expected bool) {
			dir := GinkgoT().TempDir()
			p := filepath.Join(dir, "Makefile")
			Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())

			got, err := FileContainsIncludeMakefileCommon(p)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(expected))
		},
		Entry("no include directive", "all:\n\t@echo hi\n", false),
		Entry("include at start of file", "include Makefile-common\nall:\n\t@echo hi\n", true),
		Entry("include with leading whitespace", "  include   Makefile-common\nall:\n\t@echo hi\n", true),
		Entry("commented out include", "# include Makefile-common\nall:\n\t@echo hi\n", false),
		Entry("no include in multi-target file", "foo:\n\t@echo foo\n# comment\nbar:\n\t@echo bar\n", false),
		Entry("include in the middle of the file",
			strings.Join([]string{"foo:", "\t@echo foo", "include Makefile-common", "bar:", "\t@echo bar", ""}, "\n"), true),
	)
})

var _ = Describe("PrependLineToFile", func() {
	It("should prepend a line and preserve file mode", func() {
		dir := GinkgoT().TempDir()
		p := filepath.Join(dir, "Makefile")
		original := "all:\n\t@echo hi\n"
		mode := os.FileMode(0o600)
		writeFileWithMode(p, original, mode)

		line := "include Makefile-common"
		Expect(PrependLineToFile(p, line)).To(Succeed())

		data, err := os.ReadFile(p)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal(line + "\n" + original))

		info, err := os.Stat(p)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(mode.Perm()))
	})
})

var _ = Describe("WriteYAMLWithIndent", func() {
	It("should use 2-space indentation", func() {
		dir := GinkgoT().TempDir()
		p := filepath.Join(dir, "test.yaml")

		type NestedStruct struct {
			Field1 string `yaml:"field1"`
			Field2 int    `yaml:"field2"`
		}
		type TestStruct struct {
			Name   string       `yaml:"name"`
			Nested NestedStruct `yaml:"nested"`
			Items  []string     `yaml:"items"`
		}

		data := TestStruct{
			Name: "test",
			Nested: NestedStruct{
				Field1: "value1",
				Field2: 42,
			},
			Items: []string{"item1", "item2", "item3"},
		}

		Expect(WriteYAMLWithIndent(data, p)).To(Succeed())

		content, err := os.ReadFile(p)
		Expect(err).NotTo(HaveOccurred())
		contentStr := string(content)

		Expect(contentStr).To(ContainSubstring("  field1: value1"))
		Expect(contentStr).To(ContainSubstring("  field2: 42"))
		Expect(contentStr).NotTo(ContainSubstring("    field1"))
		Expect(contentStr).NotTo(ContainSubstring("    field2"))

		// Verify round-trip
		var decoded TestStruct
		Expect(yaml.Unmarshal(content, &decoded)).To(Succeed())
		Expect(decoded.Name).To(Equal(data.Name))
		Expect(decoded.Nested.Field1).To(Equal(data.Nested.Field1))
		Expect(decoded.Nested.Field2).To(Equal(data.Nested.Field2))

		// Verify file permissions
		info, err := os.Stat(p)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o644).Perm()))
	})
})
