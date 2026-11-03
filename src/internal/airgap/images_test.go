package airgap_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/validatedpatterns/patternizer/internal/airgap"
)

var _ = Describe("DiscoverImagesFromChartDir", func() {
	It("keeps RH vault image and skips disabled injector defaults", func() {
		tmp, err := os.MkdirTemp("", "vault-values-")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmp)

		Expect(os.WriteFile(filepath.Join(tmp, "Chart.yaml"), []byte(`
apiVersion: v2
name: hashicorp-vault
version: 0.1.8
`), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "templates", "dummy.yaml"), []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: dummy
`), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "values.yaml"), []byte(`
vault:
  server:
    image:
      repository: "registry.connect.redhat.com/hashicorp/vault"
      tag: "1.21.4-ubi"
  injector:
    enabled: false
    image:
      repository: "hashicorp/vault-k8s"
      tag: "1.7.2"
  csi:
    enabled: false
    image:
      repository: "hashicorp/vault-csi-provider"
      tag: "1.7.0"
`), 0o644)).To(Succeed())

		Expect(os.MkdirAll(filepath.Join(tmp, "charts", "vault"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "charts", "vault", "Chart.yaml"), []byte(`
apiVersion: v2
name: vault
version: 0.30.0
`), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "charts", "vault", "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "charts", "vault", "templates", "dummy.yaml"), []byte("{}\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "charts", "vault", "values.yaml"), []byte(`
server:
  image:
    repository: hashicorp/vault
    tag: 1.21.2
injector:
  enabled: true
  image:
    repository: hashicorp/vault-k8s
    tag: 1.7.2
`), 0o644)).To(Succeed())

		imgs, _, err := airgap.DiscoverImagesFromChartDir(tmp)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).To(ContainElement("registry.connect.redhat.com/hashicorp/vault:1.21.4-ubi"))
		Expect(imgs).NotTo(ContainElement("hashicorp/vault-k8s:1.7.2"))
		Expect(imgs).NotTo(ContainElement("hashicorp/vault-csi-provider:1.7.0"))
		Expect(imgs).NotTo(ContainElement("hashicorp/vault:1.21.2"))
	})
})

var _ = Describe("collectImageFields registry composition", func() {
	It("joins registry/repository:tag from values when repository is complete", func() {
		tmp, err := os.MkdirTemp("", "registry-compose-")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmp)
		Expect(os.WriteFile(filepath.Join(tmp, "Chart.yaml"), []byte("apiVersion: v2\nname: t\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "templates", "x.yaml"), []byte("{}\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "values.yaml"), []byte(`
image:
  registry: quay.io
  repository: kiali/demo_travels_mysqldb
  tag: v1
`), 0o644)).To(Succeed())

		imgs, _, err := airgap.DiscoverImagesFromChartDir(tmp)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).To(ContainElement("quay.io/kiali/demo_travels_mysqldb:v1"))
	})

	It("joins registry/repository/name:tag when name is a separate field", func() {
		tmp, err := os.MkdirTemp("", "registry-name-")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmp)
		Expect(os.WriteFile(filepath.Join(tmp, "Chart.yaml"), []byte("apiVersion: v2\nname: t\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "templates", "x.yaml"), []byte("{}\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "values.yaml"), []byte(`
image:
  registry: quay.io
  repository: kiali
  name: demo_travels_control
  tag: v1
`), 0o644)).To(Succeed())

		imgs, _, err := airgap.DiscoverImagesFromChartDir(tmp)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).To(ContainElement("quay.io/kiali/demo_travels_control:v1"))
	})

	It("does not invent incomplete org-only pullspecs from values", func() {
		tmp, err := os.MkdirTemp("", "registry-org-only-")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmp)
		Expect(os.WriteFile(filepath.Join(tmp, "Chart.yaml"), []byte("apiVersion: v2\nname: t\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "templates", "x.yaml"), []byte("{}\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "values.yaml"), []byte(`
image:
  registry: quay.io
  repository: kiali
  tag: v1
`), 0o644)).To(Succeed())

		imgs, _, err := airgap.DiscoverImagesFromChartDir(tmp)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).NotTo(ContainElement("quay.io/kiali:v1"))
		Expect(imgs).To(BeEmpty())
	})

	It("resolves travelops-style images where the name is hardcoded in the template", func() {
		tmp, err := os.MkdirTemp("", "travel-control-")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmp)
		Expect(os.WriteFile(filepath.Join(tmp, "Chart.yaml"), []byte("apiVersion: v2\nname: travel-control\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "templates", "deployment.yaml"), []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: control
spec:
  template:
    spec:
      containers:
      {{- with $img := .Values.image }}
        - name: control
          image: {{ $img.registry }}/{{ $img.repository }}/demo_travels_control:{{ $img.tag }}
      {{- end }}
`), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "values.yaml"), []byte(`
image:
  registry: quay.io
  repository: kiali
  tag: v1
`), 0o644)).To(Succeed())

		imgs, _, err := airgap.DiscoverImagesFromChartDir(tmp)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).To(ContainElement("quay.io/kiali/demo_travels_control:v1"))
		Expect(imgs).NotTo(ContainElement("quay.io/kiali:v1"))
	})

	It("resolves ranged app image names from helm templates", func() {
		tmp, err := os.MkdirTemp("", "travel-agency-")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmp)
		Expect(os.WriteFile(filepath.Join(tmp, "Chart.yaml"), []byte("apiVersion: v2\nname: travel-agency\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "templates", "deployment.yaml"), []byte(`
{{- range $apps := .Values.apps }}
{{- with $img := $.Values.image }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $apps.name }}
spec:
  template:
    spec:
      containers:
        - name: {{ $apps.name }}
          image: {{ $img.registry }}/{{ $img.repository }}/demo_travels_{{ $apps.name }}:{{ $img.tag }}
{{- end }}
{{- end }}
`), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "values.yaml"), []byte(`
image:
  registry: quay.io
  repository: kiali
  tag: v1
apps:
  - name: cars
  - name: hotels
`), 0o644)).To(Succeed())

		imgs, _, err := airgap.DiscoverImagesFromChartDir(tmp)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).To(ContainElement("quay.io/kiali/demo_travels_cars:v1"))
		Expect(imgs).To(ContainElement("quay.io/kiali/demo_travels_hotels:v1"))
		Expect(imgs).NotTo(ContainElement("quay.io/kiali:v1"))
	})
})

var _ = Describe("isVPSourceImage filtering", func() {
	It("accepts VP registries and rejects docker hub short names", func() {
		tmp, err := os.MkdirTemp("", "filter-")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmp)
		Expect(os.WriteFile(filepath.Join(tmp, "Chart.yaml"), []byte("apiVersion: v2\nname: t\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "templates", "x.yaml"), []byte("{}\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "values.yaml"), []byte(`
a:
  image:
    repository: registry.redhat.io/ubi9/ubi-minimal
    tag: latest
b:
  image:
    repository: docker.io/library/nginx
    tag: latest
c:
  image:
    repository: hashicorp/vault
    tag: 1.21.2
d:
  image:
    repository: quay.io/validatedpatterns/imperative-container
    tag: v1
e:
  image:
    repository: ghcr.io/external-secrets/external-secrets
    tag: v0.10.2-ubi
`), 0o644)).To(Succeed())

		imgs, _, err := airgap.DiscoverImagesFromChartDir(tmp)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).To(ContainElement("registry.redhat.io/ubi9/ubi-minimal:latest"))
		Expect(imgs).To(ContainElement("quay.io/validatedpatterns/imperative-container:v1"))
		Expect(imgs).To(ContainElement("ghcr.io/external-secrets/external-secrets:v0.10.2-ubi"))
		Expect(imgs).NotTo(ContainElement("docker.io/library/nginx:latest"))
		Expect(imgs).NotTo(ContainElement("hashicorp/vault:1.21.2"))
	})

	It("ignores Kubernetes Service DNS image addresses", func() {
		tmp, err := os.MkdirTemp("", "k8s-svc-")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmp)
		Expect(os.WriteFile(filepath.Join(tmp, "Chart.yaml"), []byte("apiVersion: v2\nname: t\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmp, "templates"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "templates", "x.yaml"), []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: x
spec:
  containers:
  - name: c
    image: image-registry.openshift-image-registry.svc:5000/openshift/cli
`), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmp, "values.yaml"), []byte(`
image:
  repository: image-registry.openshift-image-registry.svc:5000/openshift/cli
  tag: latest
other:
  image: some-service.my-ns.svc.cluster.local:5000/ns/app:v1
keep:
  image:
    repository: quay.io/validatedpatterns/imperative-container
    tag: v1
`), 0o644)).To(Succeed())

		imgs, _, err := airgap.DiscoverImagesFromChartDir(tmp)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).To(ContainElement("quay.io/validatedpatterns/imperative-container:v1"))
		Expect(imgs).NotTo(ContainElement("image-registry.openshift-image-registry.svc:5000/openshift/cli"))
		Expect(imgs).NotTo(ContainElement("image-registry.openshift-image-registry.svc:5000/openshift/cli:latest"))
		Expect(imgs).NotTo(ContainElement("some-service.my-ns.svc.cluster.local:5000/ns/app:v1"))
	})
})
