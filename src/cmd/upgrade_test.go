package cmd_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func cloneMCGWithCommon(dir string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), "Could not init git repo")

	cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/validatedpatterns/multicloud-gitops.git")
	cmd.Dir = dir
	session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), "Could not add origin to git repo")

	cmd = exec.Command("git", "fetch", "--depth=1", "origin", "02954705e3d58e4823cd195beb8c31418f730830")
	cmd.Dir = dir
	session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), "Could not fetch SHA for last MCG commit before patternizer was run")

	cmd = exec.Command("git", "checkout", "02954705e3d58e4823cd195beb8c31418f730830")
	cmd.Dir = dir
	session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), "Could not checkout SHA for last MCG commit before patternizer was run")
}

var _ = Describe("patternizer upgrade", func() {
	Context("on a repo using common", Ordered, func() {
		var tempDir, oldMakefile string

		BeforeAll(func() {
			tempDir = createTestDir()
			cloneMCGWithCommon(tempDir)
			f, err := os.ReadFile(filepath.Join(tempDir, "Makefile"))
			Expect(err).NotTo(HaveOccurred())
			oldMakefile = string(f)
			_ = runCLI(tempDir, "upgrade")
		})

		It("should copy the common scaffold files (except Makefile)", func() {
			verifyPattenShCopied(tempDir)
			verifyMakefileCommonCopied(tempDir)
			verifyAnsibleCfgCopied(tempDir)
		})

		It("should install skills for Claude and Cursor", func() {
			verifySkillsInstalled(tempDir)
		})

		It("should inject the include for Makefile-common into the existing Makefile", func() {
			f, err := os.ReadFile(filepath.Join(tempDir, "Makefile"))
			Expect(err).NotTo(HaveOccurred())
			actualMakefile := string(f)
			Expect(strings.Contains(actualMakefile, oldMakefile)).To(BeTrue(), "Could not find contents of existing Makefile in updated Makefile")
			Expect(strings.Contains(actualMakefile, "include Makefile-common\n")).To(BeTrue(), "Could not find include for Makefile-common in updated Makefile")
		})

		It("should remove the old common directory", func() {
			_, err := os.Stat(filepath.Join(tempDir, "common"))
			Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue(), "Common directory should have been removed but was not")
		})
	})

	Context("when run multiple times", Ordered, func() {
		var tempDir, expectedMakefile string

		BeforeAll(func() {
			tempDir = createTestDir()
			cloneMCGWithCommon(tempDir)
			_ = runCLI(tempDir, "upgrade")
			f, err := os.ReadFile(filepath.Join(tempDir, "Makefile"))
			Expect(err).NotTo(HaveOccurred())
			expectedMakefile = string(f)
			_ = runCLI(tempDir, "upgrade")
		})

		It("should not update Makefiles that already include Makefile-common", func() {
			f, err := os.ReadFile(filepath.Join(tempDir, "Makefile"))
			Expect(err).NotTo(HaveOccurred())
			actualMakefile := string(f)
			Expect(actualMakefile).To(Equal(expectedMakefile), "Makefiles that already contain the include for Makefile-common should not be updated")
		})
	})
})

var _ = Describe("patternizer update (upgrade alias)", func() {
	Context("on a repo using common", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			cloneMCGWithCommon(tempDir)
			_ = runCLI(tempDir, "update")
		})

		It("should produce the same result as upgrade", func() {
			verifyPattenShCopied(tempDir)
			verifyMakefileCommonCopied(tempDir)
			verifyAnsibleCfgCopied(tempDir)
		})
	})
})

var _ = Describe("patternizer upgrade deprecation warning", func() {
	const legacyGlobal = "global:\n  pattern: test-pattern\nmain:\n  clusterGroupName: hub\n"
	const variantGlobal = "global:\n  pattern: test-pattern\nmain:\n  variant: hub\n"

	Context("when values-global.yaml uses the deprecated main.clusterGroupName", func() {
		var tempDir string
		var session *gexec.Session

		BeforeEach(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-global.yaml"), []byte(legacyGlobal), 0o644)).To(Succeed())
			session = runCLI(tempDir, "upgrade")
		})

		It("should print a deprecation warning mentioning main.variant", func() {
			out := string(session.Out.Contents())
			Expect(out).To(ContainSubstring("main.clusterGroupName"))
			Expect(out).To(ContainSubstring("main.variant"))
			Expect(out).To(ContainSubstring("0.0.78"))
		})

		It("should leave values-global.yaml unchanged", func() {
			f, err := os.ReadFile(filepath.Join(tempDir, "values-global.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(f)).To(Equal(legacyGlobal))
		})
	})

	Context("when values-global.yaml already uses main.variant and the variants folder", func() {
		var tempDir string
		var session *gexec.Session

		BeforeEach(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-global.yaml"), []byte(variantGlobal), 0o644)).To(Succeed())
			variantDir := filepath.Join(tempDir, "variants", "hub")
			Expect(os.MkdirAll(variantDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(variantDir, "values-hub.yaml"), []byte("clusterGroup:\n  name: hub\n"), 0o644)).To(Succeed())
			session = runCLI(tempDir, "upgrade")
		})

		It("should not print any deprecation warning", func() {
			out := string(session.Out.Contents())
			Expect(out).NotTo(ContainSubstring("Warning:"))
		})
	})
})

var _ = Describe("patternizer upgrade variants folder warning", func() {
	const variantGlobal = "global:\n  pattern: test-pattern\nmain:\n  variant: hub\n"

	Context("when the variants folder file is missing", func() {
		var tempDir string
		var session *gexec.Session

		BeforeEach(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-global.yaml"), []byte(variantGlobal), 0o644)).To(Succeed())
			session = runCLI(tempDir, "upgrade")
		})

		It("should print a variants folder structure warning linking to the blog", func() {
			out := string(session.Out.Contents())
			Expect(out).To(ContainSubstring("variants folder structure"))
			Expect(out).To(ContainSubstring("variants/hub/values-hub.yaml"))
			Expect(out).To(ContainSubstring("0.0.78"))
			Expect(out).To(ContainSubstring("https://validatedpatterns.io/blog/2026-08-04-variants-folder-structure/"))
		})
	})

	Context("when the variants folder file exists", func() {
		var tempDir string
		var session *gexec.Session

		BeforeEach(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-global.yaml"), []byte(variantGlobal), 0o644)).To(Succeed())
			variantDir := filepath.Join(tempDir, "variants", "hub")
			Expect(os.MkdirAll(variantDir, 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(variantDir, "values-hub.yaml"), []byte("clusterGroup:\n  name: hub\n"), 0o644)).To(Succeed())
			session = runCLI(tempDir, "upgrade")
		})

		It("should not print a variants folder structure warning", func() {
			out := string(session.Out.Contents())
			Expect(out).NotTo(ContainSubstring("variants folder structure"))
		})
	})
})

var _ = Describe("patternizer upgrade --replace-makefile", func() {
	Context("on a repo using common", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			cloneMCGWithCommon(tempDir)
			_ = runCLI(tempDir, "upgrade", "--replace-makefile")
		})

		It("should update the common scaffold files (including Makefile)", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should remove the old common directory", func() {
			_, err := os.Stat(filepath.Join(tempDir, "common"))
			Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue(), "Common directory should have been removed but was not")
		})
	})
})
