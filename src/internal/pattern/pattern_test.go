package pattern

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/validatedpatterns/patternizer/internal/types"
)

// getNestedValue is a helper function to get nested values from a map using a path.
func getNestedValue(m map[string]interface{}, path []string) interface{} {
	current := m
	for i, key := range path {
		if i == len(path)-1 {
			return current[key]
		}
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return nil
		}
		current = next
	}
	return nil
}

var _ = Describe("ProcessGlobalValues", func() {
	Context("with an existing values file containing custom fields", func() {
		var (
			tempDir    string
			valuesPath string
		)

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "pattern-test-*")
			Expect(err).NotTo(HaveOccurred())

			initialValues := map[string]interface{}{
				"global": map[string]interface{}{
					"pattern":     "existing-pattern",
					"customField": "customValue",
					"nestedCustom": map[string]interface{}{
						"key1": "value1",
						"key2": 42,
					},
				},
				"main": map[string]interface{}{
					"clusterGroupName": "custom-cluster-group",
					"multiSourceConfig": map[string]interface{}{
						"enabled":                  false,
						"clusterGroupChartVersion": "1.0.*",
						"customMultiSource":        "customValue",
					},
					"customMainField": "mainCustomValue",
				},
				"customTopLevel": map[string]interface{}{
					"someKey":    "someValue",
					"anotherKey": []string{"item1", "item2"},
				},
			}

			valuesPath = filepath.Join(tempDir, "values-global.yaml")
			initialYaml, err := yaml.Marshal(initialValues)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(valuesPath, initialYaml, 0o644)).To(Succeed())
		})

		AfterEach(func() {
			os.RemoveAll(tempDir)
		})

		It("should preserve all custom fields", func() {
			actualPatternName, clusterGroupName, err := ProcessGlobalValues("new-pattern", tempDir, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualPatternName).To(Equal("existing-pattern"))
			Expect(clusterGroupName).To(Equal("custom-cluster-group"))

			processedData, err := os.ReadFile(valuesPath)
			Expect(err).NotTo(HaveOccurred())

			var processedValues map[string]interface{}
			Expect(yaml.Unmarshal(processedData, &processedValues)).To(Succeed())

			tests := []struct {
				path     []string
				expected interface{}
			}{
				{[]string{"global", "pattern"}, "existing-pattern"},
				{[]string{"global", "customField"}, "customValue"},
				{[]string{"global", "nestedCustom", "key1"}, "value1"},
				{[]string{"global", "nestedCustom", "key2"}, 42},
				{[]string{"main", "clusterGroupName"}, "custom-cluster-group"},
				{[]string{"main", "multiSourceConfig", "enabled"}, false},
				{[]string{"main", "multiSourceConfig", "clusterGroupChartVersion"}, "1.0.*"},
				{[]string{"main", "multiSourceConfig", "customMultiSource"}, "customValue"},
				{[]string{"main", "customMainField"}, "mainCustomValue"},
				{[]string{"customTopLevel", "someKey"}, "someValue"},
			}

			for _, tt := range tests {
				Expect(getNestedValue(processedValues, tt.path)).To(Equal(tt.expected),
					"Field %v should be %v", tt.path, tt.expected)
			}

			// Verify array field is preserved
			customTopLevel, ok := processedValues["customTopLevel"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "customTopLevel should be a map")
			anotherKey, ok := customTopLevel["anotherKey"].([]interface{})
			Expect(ok).To(BeTrue(), "anotherKey should be an array")
			Expect(anotherKey).To(Equal([]interface{}{"item1", "item2"}))
		})
	})

	Context("when no existing file exists", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "pattern-test-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tempDir)
		})

		It("should create the file with defaults", func() {
			actualPatternName, clusterGroupName, err := ProcessGlobalValues("test-pattern", tempDir, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualPatternName).To(Equal("test-pattern"))
			Expect(clusterGroupName).To(Equal("prod"))

			valuesPath := filepath.Join(tempDir, "values-global.yaml")
			Expect(valuesPath).To(BeAnExistingFile())

			data, err := os.ReadFile(valuesPath)
			Expect(err).NotTo(HaveOccurred())

			var values types.ValuesGlobal
			Expect(yaml.Unmarshal(data, &values)).To(Succeed())

			Expect(values.Global.Pattern).To(Equal("test-pattern"))
			Expect(values.Main.ClusterGroupName).To(Equal("prod"))
			Expect(values.Main.MultiSourceConfig.Enabled).To(BeTrue())
			Expect(values.Main.MultiSourceConfig.ClusterGroupChartVersion).To(Equal("0.9.*"))
			Expect(values.Global.SecretLoader.Disabled).To(BeTrue())
		})
	})

	Context("with secrets enabled", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "pattern-test-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tempDir)
		})

		It("should set SecretLoader.Disabled to false", func() {
			actualPatternName, clusterGroupName, err := ProcessGlobalValues("test-pattern", tempDir, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualPatternName).To(Equal("test-pattern"))
			Expect(clusterGroupName).To(Equal("prod"))

			valuesPath := filepath.Join(tempDir, "values-global.yaml")
			Expect(valuesPath).To(BeAnExistingFile())

			data, err := os.ReadFile(valuesPath)
			Expect(err).NotTo(HaveOccurred())

			var values types.ValuesGlobal
			Expect(yaml.Unmarshal(data, &values)).To(Succeed())

			Expect(values.Global.Pattern).To(Equal("test-pattern"))
			Expect(values.Global.SecretLoader.Disabled).To(BeFalse())
		})
	})
})

var _ = Describe("ProcessClusterGroupValues", func() {
	Context("with an existing values file containing custom fields", func() {
		var (
			tempDir    string
			valuesPath string
		)

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "pattern-test-*")
			Expect(err).NotTo(HaveOccurred())

			initialValues := map[string]interface{}{
				"clusterGroup": map[string]interface{}{
					"name":         "prod",
					"isHubCluster": true,
					"namespaces":   []interface{}{"custom-ns1", "custom-ns2"},
					"projects":     []interface{}{"custom-proj1", "custom-proj2"},
					"subscriptions": map[string]interface{}{
						"custom-operator": map[string]interface{}{
							"name":      "custom-operator",
							"namespace": "custom-namespace",
							"channel":   "stable",
							"source":    "community-operators",
						},
					},
					"applications": map[string]interface{}{
						"custom-app": map[string]interface{}{
							"name":           "custom-app",
							"namespace":      "custom-namespace",
							"project":        "custom-project",
							"path":           "custom/path",
							"customAppField": "customAppValue",
						},
					},
					"customClusterField": "customClusterValue",
				},
				"customTopLevel": map[string]interface{}{
					"customKey": "customValue",
				},
			}

			valuesPath = filepath.Join(tempDir, "values-prod.yaml")
			initialYaml, err := yaml.Marshal(initialValues)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(valuesPath, initialYaml, 0o644)).To(Succeed())
		})

		AfterEach(func() {
			os.RemoveAll(tempDir)
		})

		It("should preserve custom fields", func() {
			chartPaths := []string{"charts/app1", "charts/app2"}
			Expect(ProcessClusterGroupValues("test-pattern", "prod", tempDir, chartPaths, false)).To(Succeed())

			processedData, err := os.ReadFile(valuesPath)
			Expect(err).NotTo(HaveOccurred())

			var processedValues map[string]interface{}
			Expect(yaml.Unmarshal(processedData, &processedValues)).To(Succeed())

			tests := []struct {
				path     []string
				expected interface{}
			}{
				{[]string{"clusterGroup", "name"}, "prod"},
				{[]string{"clusterGroup", "isHubCluster"}, true},
				{[]string{"clusterGroup", "customClusterField"}, "customClusterValue"},
				{[]string{"customTopLevel", "customKey"}, "customValue"},
			}

			for _, tt := range tests {
				Expect(getNestedValue(processedValues, tt.path)).To(Equal(tt.expected),
					"Field %v should be %v", tt.path, tt.expected)
			}
		})

		It("should preserve custom application fields", func() {
			chartPaths := []string{"charts/app1", "charts/app2"}
			Expect(ProcessClusterGroupValues("test-pattern", "prod", tempDir, chartPaths, false)).To(Succeed())

			processedData, err := os.ReadFile(valuesPath)
			Expect(err).NotTo(HaveOccurred())

			var processedValues map[string]interface{}
			Expect(yaml.Unmarshal(processedData, &processedValues)).To(Succeed())

			clusterGroup := processedValues["clusterGroup"].(map[string]interface{})
			applications := clusterGroup["applications"].(map[string]interface{})

			customApp := applications["custom-app"].(map[string]interface{})
			Expect(customApp["customAppField"]).To(Equal("customAppValue"))
		})

		It("should preserve custom subscriptions", func() {
			chartPaths := []string{"charts/app1", "charts/app2"}
			Expect(ProcessClusterGroupValues("test-pattern", "prod", tempDir, chartPaths, false)).To(Succeed())

			processedData, err := os.ReadFile(valuesPath)
			Expect(err).NotTo(HaveOccurred())

			var processedValues map[string]interface{}
			Expect(yaml.Unmarshal(processedData, &processedValues)).To(Succeed())

			clusterGroup := processedValues["clusterGroup"].(map[string]interface{})
			subscriptions := clusterGroup["subscriptions"].(map[string]interface{})

			customSub := subscriptions["custom-operator"].(map[string]interface{})
			Expect(customSub["channel"]).To(Equal("stable"))
		})

		It("should add new applications while preserving existing ones", func() {
			chartPaths := []string{"charts/app1", "charts/app2"}
			Expect(ProcessClusterGroupValues("test-pattern", "prod", tempDir, chartPaths, false)).To(Succeed())

			processedData, err := os.ReadFile(valuesPath)
			Expect(err).NotTo(HaveOccurred())

			var processedValues map[string]interface{}
			Expect(yaml.Unmarshal(processedData, &processedValues)).To(Succeed())

			clusterGroup := processedValues["clusterGroup"].(map[string]interface{})
			applications := clusterGroup["applications"].(map[string]interface{})

			Expect(applications).To(HaveKey("app1"))
			Expect(applications).To(HaveKey("app2"))
			Expect(applications).To(HaveKey("custom-app"))
		})
	})
})
