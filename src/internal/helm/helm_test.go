package helm

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// createTestChartStructure creates a comprehensive test directory structure for helm chart testing.
func createTestChartStructure() string {
	tempDir, err := os.MkdirTemp("", "helm-test-*")
	Expect(err).NotTo(HaveOccurred())

	// chart1 (valid top-level chart)
	chart1Dir := filepath.Join(tempDir, "chart1")
	Expect(os.MkdirAll(filepath.Join(chart1Dir, "templates"), 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(chart1Dir, "Chart.yaml"), []byte("name: chart1\nversion: 1.0.0\n"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(chart1Dir, "values.yaml"), []byte("# values\n"), 0o644)).To(Succeed())

	// chart2 (valid top-level chart with sub-chart)
	chart2Dir := filepath.Join(tempDir, "chart2")
	chart2SubchartDir := filepath.Join(chart2Dir, "charts", "subchart")
	Expect(os.MkdirAll(filepath.Join(chart2Dir, "templates"), 0o755)).To(Succeed())
	Expect(os.MkdirAll(filepath.Join(chart2SubchartDir, "templates"), 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(chart2Dir, "Chart.yaml"), []byte("name: chart2\nversion: 1.0.0\n"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(chart2Dir, "values.yaml"), []byte("# values\n"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(chart2SubchartDir, "Chart.yaml"), []byte("name: subchart\nversion: 1.0.0\n"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(chart2SubchartDir, "values.yaml"), []byte("# subchart values\n"), 0o644)).To(Succeed())

	// incomplete-chart (missing templates directory)
	incompleteChartDir := filepath.Join(tempDir, "incomplete-chart")
	Expect(os.MkdirAll(incompleteChartDir, 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(incompleteChartDir, "Chart.yaml"), []byte("name: incomplete\nversion: 1.0.0\n"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(incompleteChartDir, "values.yaml"), []byte("# values\n"), 0o644)).To(Succeed())

	// missing-chart-yaml
	missingChartYamlDir := filepath.Join(tempDir, "missing-chart-yaml")
	Expect(os.MkdirAll(filepath.Join(missingChartYamlDir, "templates"), 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(missingChartYamlDir, "values.yaml"), []byte("# values\n"), 0o644)).To(Succeed())

	// missing-values-yaml
	missingValuesYamlDir := filepath.Join(tempDir, "missing-values-yaml")
	Expect(os.MkdirAll(filepath.Join(missingValuesYamlDir, "templates"), 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(missingValuesYamlDir, "Chart.yaml"), []byte("name: missing-values\nversion: 1.0.0\n"), 0o644)).To(Succeed())

	// templates-is-file (templates is a file, not directory)
	templatesIsFileDir := filepath.Join(tempDir, "templates-is-file")
	Expect(os.MkdirAll(templatesIsFileDir, 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(templatesIsFileDir, "Chart.yaml"), []byte("name: templates-file\nversion: 1.0.0\n"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(templatesIsFileDir, "values.yaml"), []byte("# values\n"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(templatesIsFileDir, "templates"), []byte("not a directory\n"), 0o644)).To(Succeed())

	// hidden chart
	hiddenChartDir := filepath.Join(tempDir, ".hidden-chart")
	Expect(os.MkdirAll(filepath.Join(hiddenChartDir, "templates"), 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(hiddenChartDir, "Chart.yaml"), []byte("name: hidden\nversion: 1.0.0\n"), 0o644)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(hiddenChartDir, "values.yaml"), []byte("# values\n"), 0o644)).To(Succeed())

	// not-a-chart directory
	notChartDir := filepath.Join(tempDir, "not-a-chart")
	Expect(os.MkdirAll(notChartDir, 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(notChartDir, "some-file.txt"), []byte("not a chart\n"), 0o644)).To(Succeed())

	return tempDir
}

var _ = Describe("FindTopLevelCharts", func() {
	It("should only return top-level charts and skip sub-charts and non-chart directories", func() {
		tempDir := createTestChartStructure()
		defer os.RemoveAll(tempDir)

		charts, err := FindTopLevelCharts(tempDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(charts).To(HaveLen(2))
		Expect(charts).To(ContainElements("chart1", "chart2"))

		// Verify unexpected charts are not present
		Expect(charts).NotTo(ContainElement("subchart"))
		Expect(charts).NotTo(ContainElement("charts/subchart"))
		Expect(charts).NotTo(ContainElement("incomplete-chart"))
		Expect(charts).NotTo(ContainElement(".hidden-chart"))
		Expect(charts).NotTo(ContainElement("not-a-chart"))
	})
})

var _ = Describe("IsHelmChart", func() {
	var tempDir string

	BeforeEach(func() {
		tempDir = createTestChartStructure()
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	DescribeTable("should correctly identify helm charts",
		func(chartDir string, expected bool) {
			testDir := filepath.Join(tempDir, chartDir)
			Expect(IsHelmChart(testDir)).To(Equal(expected))
		},
		Entry("valid helm chart 1", "chart1", true),
		Entry("valid helm chart 2", "chart2", true),
		Entry("valid subchart (still valid helm chart)", "chart2/charts/subchart", true),
		Entry("hidden chart (still valid helm chart)", ".hidden-chart", true),
		Entry("incomplete chart - missing templates", "incomplete-chart", false),
		Entry("missing Chart.yaml", "missing-chart-yaml", false),
		Entry("missing values.yaml", "missing-values-yaml", false),
		Entry("templates is a file not directory", "templates-is-file", false),
		Entry("not a chart directory", "not-a-chart", false),
	)
})
