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

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common scaffold files (except Makefile)", func() {
			verifyPattenShCopied(tempDir)
			verifyMakefileCommonCopied(tempDir)
			verifyAnsibleCfgCopied(tempDir)
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

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should not update Makefiles that already include Makefile-common", func() {
			f, err := os.ReadFile(filepath.Join(tempDir, "Makefile"))
			Expect(err).NotTo(HaveOccurred())
			actualMakefile := string(f)
			Expect(actualMakefile).To(Equal(expectedMakefile), "Makefiles that already contain the include for Makefile-common should not be updated")
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

		AfterAll(func() {
			os.RemoveAll(tempDir)
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
