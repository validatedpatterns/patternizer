package pattern

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GeneratePatternMetadata", func() {
	Context("on a directory without a git remote", func() {
		var tempDir string

		BeforeEach(func() {
			tempDir = GinkgoT().TempDir()
		})

		It("should create pattern-metadata.yaml with fallback values", func() {
			Expect(GeneratePatternMetadata("my-pattern", tempDir)).To(Succeed())

			metadataPath := filepath.Join(tempDir, "pattern-metadata.yaml")
			Expect(metadataPath).To(BeAnExistingFile())

			content, err := os.ReadFile(metadataPath)
			Expect(err).NotTo(HaveOccurred())

			s := string(content)
			Expect(s).To(ContainSubstring("name: my-pattern"))
			Expect(s).To(ContainSubstring("display_name: My Pattern"))
			Expect(s).To(ContainSubstring("org: CHANGEME"))
			Expect(s).To(ContainSubstring("repo_url: https://github.com/CHANGEME/my-pattern"))
			Expect(s).To(ContainSubstring("issues_url: https://github.com/CHANGEME/my-pattern/issues"))
			Expect(s).To(ContainSubstring("tier: sandbox"))
			Expect(s).To(ContainSubstring("# requirements:"))
			Expect(s).To(ContainSubstring("extra_features:"))
		})
	})

	Context("on a directory with a git remote", func() {
		var tempDir string

		BeforeEach(func() {
			tempDir = GinkgoT().TempDir()
			cmd := exec.Command("git", "init", tempDir)
			Expect(cmd.Run()).To(Succeed())
			cmd = exec.Command("git", "-C", tempDir, "remote", "add", "origin", "https://github.com/testorg/my-pattern.git")
			Expect(cmd.Run()).To(Succeed())
		})

		It("should detect org and repo URL from git remote", func() {
			Expect(GeneratePatternMetadata("my-pattern", tempDir)).To(Succeed())

			content, err := os.ReadFile(filepath.Join(tempDir, "pattern-metadata.yaml"))
			Expect(err).NotTo(HaveOccurred())

			s := string(content)
			Expect(s).To(ContainSubstring("org: testorg"))
			Expect(s).To(ContainSubstring("repo_url: https://github.com/testorg/my-pattern"))
			Expect(s).To(ContainSubstring("issues_url: https://github.com/testorg/my-pattern/issues"))
		})
	})

	Context("when pattern-metadata.yaml already exists", func() {
		var tempDir string

		BeforeEach(func() {
			tempDir = GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "pattern-metadata.yaml"), []byte("existing content"), 0o644)).To(Succeed())
		})

		It("should not overwrite the existing file", func() {
			Expect(GeneratePatternMetadata("my-pattern", tempDir)).To(Succeed())

			content, err := os.ReadFile(filepath.Join(tempDir, "pattern-metadata.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("existing content"))
		})
	})
})

var _ = Describe("parseGitRemoteURL", func() {
	DescribeTable("should parse various remote URL formats",
		func(raw, expectedOrg, expectedURL string) {
			org, repoURL := parseGitRemoteURL(raw, "fallback-org", "https://fallback.example.com")
			Expect(org).To(Equal(expectedOrg))
			Expect(repoURL).To(Equal(expectedURL))
		},
		Entry("HTTPS URL", "https://github.com/validatedpatterns/multicloud-gitops.git",
			"validatedpatterns", "https://github.com/validatedpatterns/multicloud-gitops"),
		Entry("HTTPS URL without .git", "https://github.com/validatedpatterns/multicloud-gitops",
			"validatedpatterns", "https://github.com/validatedpatterns/multicloud-gitops"),
		Entry("SSH URL", "git@github.com:validatedpatterns/multicloud-gitops.git",
			"validatedpatterns", "https://github.com/validatedpatterns/multicloud-gitops"),
	)

	It("should fall back on invalid URLs", func() {
		org, repoURL := parseGitRemoteURL("not-a-url", "fallback-org", "https://fallback.example.com")
		Expect(org).To(Equal("fallback-org"))
		Expect(repoURL).To(Equal("https://fallback.example.com"))
	})
})

var _ = Describe("toDisplayName", func() {
	DescribeTable("should convert kebab-case to title case",
		func(input, expected string) {
			Expect(toDisplayName(input)).To(Equal(expected))
		},
		Entry("single word", "pattern", "Pattern"),
		Entry("two words", "multicloud-gitops", "Multicloud Gitops"),
		Entry("three words", "my-cool-pattern", "My Cool Pattern"),
	)
})

var _ = Describe("detectGitRemote", func() {
	It("should return fallback values for a non-git directory", func() {
		tempDir := GinkgoT().TempDir()
		org, repoURL := detectGitRemote("test-pattern", tempDir)
		Expect(org).To(Equal("CHANGEME"))
		Expect(repoURL).To(Equal("https://github.com/CHANGEME/test-pattern"))
	})

	It("should detect remote from a git repo", func() {
		tempDir := GinkgoT().TempDir()
		Expect(exec.Command("git", "init", tempDir).Run()).To(Succeed())
		Expect(exec.Command("git", "-C", tempDir, "remote", "add", "origin", "git@github.com:myorg/myrepo.git").Run()).To(Succeed())

		org, repoURL := detectGitRemote("myrepo", tempDir)
		Expect(org).To(Equal("myorg"))
		Expect(strings.HasPrefix(repoURL, "https://github.com/myorg/myrepo")).To(BeTrue())
	})
})
