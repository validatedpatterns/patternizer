package cmd_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/validatedpatterns/patternizer/internal/types"
)

const customGlobalValues = `
global:
  pattern: test-pattern
main:
  clusterGroupName: test
`

const customClusterGroupValues = `
clusterGroup:
  name: test

  customClusterField: user-cluster-config

  applications:
    custom-user-app:
      name: custom-user-app
      namespace: user-namespace
      path: user/path
      customAppField: user-app-config
      project: custom-pattern-name

customClusterTopLevel: user-cluster-top-level
`

const customSecretTemplate = `
version: "2.0"

secrets:
  - name: customSecret
    fields:
    - name: test
      value: test
`

var _ = Describe("patternizer init", func() {
	Context("on an empty directory", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			_ = runCLI(tempDir, "init")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common pattern scaffold files", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should create an appropriate global values file", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: filepath.Base(tempDir),
					SecretLoader: types.SecretLoader{
						Disabled: true,
					},
				},
				Main: types.Main{
					ClusterGroupName: "prod",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should create an appropriate clustergroup values file", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-prod.yaml")
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name:          "prod",
					Namespaces:    []types.NamespaceEntry{types.NewNamespaceEntry(filepath.Base(tempDir))},
					Subscriptions: map[string]types.Subscription{},
					Applications:  map[string]types.Application{},
				},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})

	Context("on a directory containing helm charts", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			addDummyChart(tempDir, "test-app1")
			addDummyChart(tempDir, "test-app2")
			_ = runCLI(tempDir, "init")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common pattern scaffold files", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should create an appropriate global values file", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: filepath.Base(tempDir),
					SecretLoader: types.SecretLoader{
						Disabled: true,
					},
				},
				Main: types.Main{
					ClusterGroupName: "prod",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should create an appropriate clustergroup values file", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-prod.yaml")
			expectedNamespace := filepath.Base(tempDir)
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name:          "prod",
					Namespaces:    []types.NamespaceEntry{types.NewNamespaceEntry(expectedNamespace)},
					Subscriptions: map[string]types.Subscription{},
					Applications: map[string]types.Application{
						"test-app1": {
							Name:      "test-app1",
							Namespace: expectedNamespace,
							Path:      "charts/test-app1",
						},
						"test-app2": {
							Name:      "test-app2",
							Namespace: expectedNamespace,
							Path:      "charts/test-app2",
						},
					},
				},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})

	Context("on a directory with a custom global values file", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-global.yaml"), []byte(customGlobalValues), 0o644)).To(Succeed())
			_ = runCLI(tempDir, "init")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common pattern scaffold files", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should create an appropriate global values file", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: "test-pattern",
					SecretLoader: types.SecretLoader{
						Disabled: true,
					},
				},
				Main: types.Main{
					ClusterGroupName: "test",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should create an appropriate clustergroup values file", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-test.yaml")
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name:          "test",
					Namespaces:    []types.NamespaceEntry{types.NewNamespaceEntry("test-pattern")},
					Subscriptions: map[string]types.Subscription{},
					Applications:  map[string]types.Application{},
				},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})

	Context("on a directory with custom global values and clustergroup values files", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-global.yaml"), []byte(customGlobalValues), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(tempDir, "values-test.yaml"), []byte(customClusterGroupValues), 0o644)).To(Succeed())
			_ = runCLI(tempDir, "init")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common pattern scaffold files", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should create an appropriate global values file", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: "test-pattern",
					SecretLoader: types.SecretLoader{
						Disabled: true,
					},
				},
				Main: types.Main{
					ClusterGroupName: "test",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should create an appropriate clustergroup values file", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-test.yaml")
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name:          "test",
					Namespaces:    []types.NamespaceEntry{types.NewNamespaceEntry("test-pattern")},
					Subscriptions: map[string]types.Subscription{},
					Applications: map[string]types.Application{
						"custom-user-app": {
							Name:        "custom-user-app",
							Namespace:   "user-namespace",
							Path:        "user/path",
							Project:     "custom-pattern-name",
							OtherFields: map[string]interface{}{"customAppField": "user-app-config"},
						},
					},
					OtherFields: map[string]interface{}{"customClusterField": "user-cluster-config"},
				},
				OtherFields: map[string]interface{}{"customClusterTopLevel": "user-cluster-top-level"},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})
})

var _ = Describe("patternizer init --with-secrets", func() {
	Context("on an empty directory", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			_ = runCLI(tempDir, "init", "--with-secrets")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common pattern scaffold files", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should copy the secrets template file", func() {
			verifySecretTemplateCopied(tempDir)
		})

		It("should create an appropriate global values file", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: filepath.Base(tempDir),
					SecretLoader: types.SecretLoader{
						Disabled: false,
					},
				},
				Main: types.Main{
					ClusterGroupName: "prod",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should create an appropriate clustergroup values file", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-prod.yaml")
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name: "prod",
					Namespaces: []types.NamespaceEntry{
						types.NewNamespaceEntry(filepath.Base(tempDir)),
						types.NewNamespaceEntry("vault"),
						types.NewNamespaceEntry("golang-external-secrets"),
					},
					Subscriptions: map[string]types.Subscription{},
					Applications: map[string]types.Application{
						"vault": {
							Name:         "vault",
							Namespace:    "vault",
							Chart:        "hashicorp-vault",
							ChartVersion: "0.1.*",
						},
						"golang-external-secrets": {
							Name:         "golang-external-secrets",
							Namespace:    "golang-external-secrets",
							Chart:        "golang-external-secrets",
							ChartVersion: "0.1.*",
						},
					},
				},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})

	Context("on a directory containing helm charts", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			addDummyChart(tempDir, "test-app1")
			addDummyChart(tempDir, "test-app2")
			_ = runCLI(tempDir, "init", "--with-secrets")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common pattern scaffold files", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should copy the secrets template file", func() {
			verifySecretTemplateCopied(tempDir)
		})

		It("should create an appropriate global values file", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: filepath.Base(tempDir),
					SecretLoader: types.SecretLoader{
						Disabled: false,
					},
				},
				Main: types.Main{
					ClusterGroupName: "prod",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should create an appropriate clustergroup values file", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-prod.yaml")
			expectedNamespace := filepath.Base(tempDir)
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name: "prod",
					Namespaces: []types.NamespaceEntry{
						types.NewNamespaceEntry(expectedNamespace),
						types.NewNamespaceEntry("vault"),
						types.NewNamespaceEntry("golang-external-secrets"),
					},
					Subscriptions: map[string]types.Subscription{},
					Applications: map[string]types.Application{
						"test-app1": {
							Name:      "test-app1",
							Namespace: expectedNamespace,
							Path:      "charts/test-app1",
						},
						"test-app2": {
							Name:      "test-app2",
							Namespace: expectedNamespace,
							Path:      "charts/test-app2",
						},
						"vault": {
							Name:         "vault",
							Namespace:    "vault",
							Chart:        "hashicorp-vault",
							ChartVersion: "0.1.*",
						},
						"golang-external-secrets": {
							Name:         "golang-external-secrets",
							Namespace:    "golang-external-secrets",
							Chart:        "golang-external-secrets",
							ChartVersion: "0.1.*",
						},
					},
				},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})

	Context("on a directory with a custom global values file", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-global.yaml"), []byte(customGlobalValues), 0o644)).To(Succeed())
			_ = runCLI(tempDir, "init", "--with-secrets")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common pattern scaffold files", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should copy the secrets template file", func() {
			verifySecretTemplateCopied(tempDir)
		})

		It("should create an appropriate global values file", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: "test-pattern",
					SecretLoader: types.SecretLoader{
						Disabled: false,
					},
				},
				Main: types.Main{
					ClusterGroupName: "test",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should create an appropriate clustergroup values file", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-test.yaml")
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name: "test",
					Namespaces: []types.NamespaceEntry{
						types.NewNamespaceEntry("test-pattern"),
						types.NewNamespaceEntry("vault"),
						types.NewNamespaceEntry("golang-external-secrets"),
					},
					Subscriptions: map[string]types.Subscription{},
					Applications: map[string]types.Application{
						"vault": {
							Name:         "vault",
							Namespace:    "vault",
							Chart:        "hashicorp-vault",
							ChartVersion: "0.1.*",
						},
						"golang-external-secrets": {
							Name:         "golang-external-secrets",
							Namespace:    "golang-external-secrets",
							Chart:        "golang-external-secrets",
							ChartVersion: "0.1.*",
						},
					},
				},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})

	Context("after running patternizer init", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			addDummyChart(tempDir, "test-app1")
			addDummyChart(tempDir, "test-app2")
			_ = runCLI(tempDir, "init")
			_ = runCLI(tempDir, "init", "--with-secrets")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the secrets template file", func() {
			verifySecretTemplateCopied(tempDir)
		})

		It("should update the global values file to load secrets", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: filepath.Base(tempDir),
					SecretLoader: types.SecretLoader{
						Disabled: false,
					},
				},
				Main: types.Main{
					ClusterGroupName: "prod",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should update the clustergroup values file to include secrets", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-prod.yaml")
			expectedNamespace := filepath.Base(tempDir)
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name: "prod",
					Namespaces: []types.NamespaceEntry{
						types.NewNamespaceEntry(expectedNamespace),
						types.NewNamespaceEntry("vault"),
						types.NewNamespaceEntry("golang-external-secrets"),
					},
					Subscriptions: map[string]types.Subscription{},
					Applications: map[string]types.Application{
						"test-app1": {
							Name:      "test-app1",
							Namespace: expectedNamespace,
							Path:      "charts/test-app1",
						},
						"test-app2": {
							Name:      "test-app2",
							Namespace: expectedNamespace,
							Path:      "charts/test-app2",
						},
						"vault": {
							Name:         "vault",
							Namespace:    "vault",
							Chart:        "hashicorp-vault",
							ChartVersion: "0.1.*",
						},
						"golang-external-secrets": {
							Name:         "golang-external-secrets",
							Namespace:    "golang-external-secrets",
							Chart:        "golang-external-secrets",
							ChartVersion: "0.1.*",
						},
					},
				},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})

	Context("on a directory with custom global values and clustergroup values files", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-global.yaml"), []byte(customGlobalValues), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(tempDir, "values-test.yaml"), []byte(customClusterGroupValues), 0o644)).To(Succeed())
			_ = runCLI(tempDir, "init", "--with-secrets")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should copy the common pattern scaffold files", func() {
			verifyScaffoldFilesCopied(tempDir)
		})

		It("should copy the secrets template file", func() {
			verifySecretTemplateCopied(tempDir)
		})

		It("should create an appropriate global values file", func() {
			globalValuesFile := filepath.Join(tempDir, "values-global.yaml")
			expectedGlobalValues := types.ValuesGlobal{
				Global: types.Global{
					Pattern: "test-pattern",
					SecretLoader: types.SecretLoader{
						Disabled: false,
					},
				},
				Main: types.Main{
					ClusterGroupName: "test",
					MultiSourceConfig: types.MultiSourceConfig{
						Enabled:                  true,
						ClusterGroupChartVersion: "0.9.*",
					},
				},
			}
			verifyGlobalValues(globalValuesFile, &expectedGlobalValues)
		})

		It("should create an appropriate clustergroup values file", func() {
			clusterGroupValuesFile := filepath.Join(tempDir, "values-test.yaml")
			expectedClusterGroupValues := types.ValuesClusterGroup{
				ClusterGroup: types.ClusterGroup{
					Name: "test",
					Namespaces: []types.NamespaceEntry{
						types.NewNamespaceEntry("test-pattern"),
						types.NewNamespaceEntry("vault"),
						types.NewNamespaceEntry("golang-external-secrets"),
					},
					Subscriptions: map[string]types.Subscription{},
					Applications: map[string]types.Application{
						"custom-user-app": {
							Name:        "custom-user-app",
							Namespace:   "user-namespace",
							Path:        "user/path",
							Project:     "custom-pattern-name",
							OtherFields: map[string]interface{}{"customAppField": "user-app-config"},
						},
						"vault": {
							Name:         "vault",
							Namespace:    "vault",
							Chart:        "hashicorp-vault",
							ChartVersion: "0.1.*",
						},
						"golang-external-secrets": {
							Name:         "golang-external-secrets",
							Namespace:    "golang-external-secrets",
							Chart:        "golang-external-secrets",
							ChartVersion: "0.1.*",
						},
					},
					OtherFields: map[string]interface{}{"customClusterField": "user-cluster-config"},
				},
				OtherFields: map[string]interface{}{"customClusterTopLevel": "user-cluster-top-level"},
			}
			verifyClusterGroupValues(clusterGroupValuesFile, &expectedClusterGroupValues)
		})
	})

	Context("on a directory with a custom secret template", Ordered, func() {
		var tempDir string

		BeforeAll(func() {
			tempDir = createTestDir()
			Expect(os.WriteFile(filepath.Join(tempDir, "values-secret.yaml.template"), []byte(customSecretTemplate), 0o644)).To(Succeed())
			_ = runCLI(tempDir, "init", "--with-secrets")
		})

		AfterAll(func() {
			os.RemoveAll(tempDir)
		})

		It("should not modify the secrets template file", func() {
			actual, err := os.ReadFile(filepath.Join(tempDir, "values-secret.yaml.template"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(actual)).To(Equal(customSecretTemplate))
		})
	})
})
