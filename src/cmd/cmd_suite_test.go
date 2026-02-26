package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v3"

	"github.com/validatedpatterns/patternizer/internal/types"
)

var (
	binaryPath    string
	resourcesPath string
	projectRoot   string
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}

var _ = BeforeSuite(func() {
	var err error

	wd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	projectRoot, err = filepath.Abs(filepath.Join(wd, "..", ".."))
	Expect(err).NotTo(HaveOccurred())

	resourcesPath = filepath.Join(projectRoot, "resources")
	Expect(resourcesPath).To(BeADirectory(), "Could not find resources directory")
	os.Setenv("PATTERNIZER_RESOURCES_DIR", resourcesPath)

	binaryPath, err = gexec.Build(filepath.Join(projectRoot, "src"))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func verifyPattenShCopied(dir string) {
	actual := filepath.Join(dir, "pattern.sh")
	expected := filepath.Join(resourcesPath, "pattern.sh")
	verifyFilesMatch(actual, expected)

	// verify pattern.sh is executable
	info, err := os.Stat(actual)
	Expect(err).NotTo(HaveOccurred())
	Expect(info.Mode() & 0o111).NotTo(Equal(0))
}

func verifyMakefileCommonCopied(dir string) {
	actual := filepath.Join(dir, "Makefile-common")
	expected := filepath.Join(resourcesPath, "Makefile-common")
	verifyFilesMatch(actual, expected)
}

func verifyMakefileCopied(dir string) {
	actual := filepath.Join(dir, "Makefile")
	expected := filepath.Join(resourcesPath, "Makefile")
	verifyFilesMatch(actual, expected)
}

func verifyAnsibleCfgCopied(dir string) {
	actual := filepath.Join(dir, "ansible.cfg")
	expected := filepath.Join(resourcesPath, "ansible.cfg")
	verifyFilesMatch(actual, expected)
}

func verifyScaffoldFilesCopied(dir string) {
	verifyPattenShCopied(dir)
	verifyMakefileCommonCopied(dir)
	verifyMakefileCopied(dir)
	verifyAnsibleCfgCopied(dir)
}

func verifySecretTemplateCopied(dir string) {
	actual := filepath.Join(dir, "values-secret.yaml.template")
	expected := filepath.Join(dir, "values-secret.yaml.template")
	verifyFilesMatch(actual, expected)
}

func verifyFilesMatch(file1, file2 string) {
	file1Contents, err := os.ReadFile(file1)
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Could not read file %s", file1))

	file2Contents, err := os.ReadFile(file2)
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Could not read file %s", file2))

	Expect(bytes.Equal(file1Contents, file2Contents)).To(BeTrue(), fmt.Sprintf("%s and %s have different contents", file1, file2))
}

func verifyGlobalValues(valuesFile string, expectedGlobalValues *types.ValuesGlobal) {
	f, err := os.ReadFile(valuesFile)
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Could not read file %s", valuesFile))

	var globalValues *types.ValuesGlobal
	err = yaml.Unmarshal(f, &globalValues)
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Could not unmarshal %s into the ValuesGlobal type", valuesFile))

	Expect(*globalValues).To(Equal(*expectedGlobalValues), fmt.Sprintf("Global values in %s differ from expected values", valuesFile))
}

func verifyClusterGroupValues(valuesFile string, expectedClusterGroupValues *types.ValuesClusterGroup) {
	f, err := os.ReadFile(valuesFile)
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Could not read file %s", valuesFile))

	var clusterGroupValues *types.ValuesClusterGroup
	err = yaml.Unmarshal(f, &clusterGroupValues)
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Could not unmarshal %s into the ValuesClusterGroup type", valuesFile))

	Expect(*clusterGroupValues).To(Equal(*expectedClusterGroupValues), fmt.Sprintf("Clustergroup values in %s differ from expected values", valuesFile))
}

func createTestDir() string {
	dir, err := os.MkdirTemp("", "patternizer-test")
	Expect(err).NotTo(HaveOccurred())
	return dir
}

func runCLI(dir string, args ...string) *gexec.Session {
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATTERNIZER_RESOURCES_DIR="+resourcesPath)

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
	return session
}

func addDummyChart(dir, name string) {
	path := filepath.Join(dir, "charts", name)
	Expect(os.MkdirAll(path, 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(path, "Chart.yaml"), []byte("name: "+name+"\nversion: 0.1.0"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(path, "values.yaml"), []byte("replicaCount: 1"), 0o644)).To(Succeed())
	Expect(os.MkdirAll(filepath.Join(path, "templates"), 0o755)).To(Succeed())
}
