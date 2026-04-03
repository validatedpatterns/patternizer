package main

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/validatedpatterns/patternizer/internal/fileutils"
	"github.com/validatedpatterns/patternizer/internal/types"
)

var _ = Describe("GetResourcePath", func() {
	AfterEach(func() {
		os.Unsetenv("PATTERNIZER_RESOURCES_DIR")
	})

	It("should return the path when the environment variable is set", func() {
		os.Setenv("PATTERNIZER_RESOURCES_DIR", "/tmp/test")
		path, err := fileutils.GetResourcesPath()
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal("/tmp/test"))
	})

	It("should return an error when the environment variable is unset", func() {
		os.Unsetenv("PATTERNIZER_RESOURCES_DIR")
		path, err := fileutils.GetResourcesPath()
		Expect(err).To(HaveOccurred())
		Expect(path).To(BeEmpty())
	})
})

var _ = Describe("NewDefaultValuesGlobal", func() {
	It("should create default global values with expected defaults", func() {
		values := types.NewDefaultValuesGlobal()

		Expect(values.Main.ClusterGroupName).To(Equal("prod"))
		Expect(values.Main.MultiSourceConfig.Enabled).To(BeTrue())
		Expect(values.Main.MultiSourceConfig.ClusterGroupChartVersion).To(Equal("0.9.*"))
	})
})

var _ = Describe("NewDefaultValuesClusterGroup", func() {
	Context("without secrets", func() {
		It("should create a cluster group with the correct name and namespaces", func() {
			values := types.NewDefaultValuesClusterGroup("test-pattern", "test-group", []string{"charts/app1", "charts/app2"}, false)

			Expect(values.ClusterGroup.Name).To(Equal("test-group"))
			Expect(values.ClusterGroup.Namespaces).To(HaveLen(1))
		})
	})

	Context("with secrets", func() {
		var valuesWithSecrets *types.ValuesClusterGroup

		BeforeEach(func() {
			valuesWithSecrets = types.NewDefaultValuesClusterGroup("test-pattern", "test-group", []string{"charts/app1"}, true)
		})

		It("should include all expected namespaces", func() {
			Expect(valuesWithSecrets.ClusterGroup.Namespaces).To(HaveLen(4))
		})

		It("should include the vault application", func() {
			Expect(valuesWithSecrets.ClusterGroup.Applications).To(HaveKey("vault"))
		})

		It("should include the openshift-external-secrets application", func() {
			Expect(valuesWithSecrets.ClusterGroup.Applications).To(HaveKey("openshift-external-secrets"))
		})

		It("should include the eso subscription", func() {
			Expect(valuesWithSecrets.ClusterGroup.Subscriptions).To(HaveKey("eso"))
		})
	})
})
