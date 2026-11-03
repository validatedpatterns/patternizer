package airgap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DiscoverImagesFromChartDir discovers container images required to deploy a chart.
func DiscoverImagesFromChartDir(chartDir string) (images, warnings []string, err error) {
	found := map[string]struct{}{}

	templated, warn, tmplErr := imagesFromHelmTemplate(chartDir)
	if tmplErr != nil {
		warnings = append(warnings, fmt.Sprintf("%s helm template: %v", chartDir, tmplErr))
	}
	if warn != "" {
		warnings = append(warnings, warn)
	}
	for _, img := range templated {
		found[img] = struct{}{}
	}

	fromValues, valuesErr := imagesFromValues(chartDir)
	if valuesErr != nil {
		warnings = append(warnings, fmt.Sprintf("%s values scrape: %v", chartDir, valuesErr))
	}
	for _, img := range fromValues {
		found[img] = struct{}{}
	}

	return filterRequiredImages(found, warnings), warnings, nil
}

// DiscoverOCIChartImages pulls a remote OCI Helm chart and discovers deploy-time images.
func DiscoverOCIChartImages(helmRepo, chartName, version string, skipNetwork bool) (images, warnings []string, err error) {
	if skipNetwork {
		return nil, []string{fmt.Sprintf("skipped OCI image discovery for %s (skip-oci-resolve)", chartName)}, nil
	}
	if _, err := exec.LookPath("helm"); err != nil {
		return nil, nil, fmt.Errorf("helm required to discover images for %s", chartName)
	}

	ref := fmt.Sprintf("oci://%s/%s", helmRepo, chartName)

	tmp, err := os.MkdirTemp("", "patternizer-chart-"+sanitizeDirName(chartName)+"-")
	if err != nil {
		return nil, warnings, err
	}
	defer os.RemoveAll(tmp)

	args := []string{"pull", ref, "--untar", "--untardir", tmp}
	if version != "" && version != "*" && !strings.Contains(version, "*") {
		args = append(args, "--version", version)
	}
	cmd := exec.Command("helm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		showImgs, showErr := imagesFromHelmShowValues(ref, version)
		if showErr != nil {
			return nil, warnings, fmt.Errorf("helm pull %s: %w (%s); show values: %v", ref, err, truncate(string(out), 300), showErr)
		}
		warnings = append(warnings, fmt.Sprintf("%s helm pull failed, using show values only: %v", chartName, err))
		return filterRequiredImages(sliceToSet(showImgs), warnings), warnings, nil
	}

	chartDir := findPulledChartDir(tmp, chartName)
	if chartDir == "" {
		return nil, warnings, fmt.Errorf("pulled chart dir not found for %s", chartName)
	}

	if _, err := os.Stat(filepath.Join(chartDir, "Chart.yaml")); err == nil {
		depCmd := exec.Command("helm", "dependency", "build", chartDir)
		if depOut, depErr := depCmd.CombinedOutput(); depErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s helm dependency build: %v (%s)", chartName, depErr, truncate(string(depOut), 200)))
		}
	}

	pulled, pullWarns, discErr := DiscoverImagesFromChartDir(chartDir)
	warnings = append(warnings, pullWarns...)
	if discErr != nil {
		warnings = append(warnings, fmt.Sprintf("%s discover: %v", chartName, discErr))
	}
	return pulled, warnings, nil
}

func imagesFromHelmShowValues(ref, version string) ([]string, error) {
	args := []string{"show", "values", ref}
	if version != "" && version != "*" && !strings.Contains(version, "*") {
		args = append(args, "--version", version)
	}
	cmd := exec.Command("helm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, truncate(string(out), 300))
	}
	var values map[string]interface{}
	if err := yaml.Unmarshal(out, &values); err != nil {
		return nil, err
	}
	var images []string
	collectImageFields(values, &images)
	return dropRedundantImagePrefixes(uniqueSorted(images)), nil
}

func findPulledChartDir(parent, chartName string) string {
	candidate := filepath.Join(parent, chartName)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			return filepath.Join(parent, e.Name())
		}
	}
	return ""
}

func imagesFromValues(chartDir string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(chartDir, "values.yaml"))
	if err != nil {
		return nil, err
	}
	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, err
	}
	var images []string
	collectImageFields(values, &images)
	return images, nil
}

func collectImageFields(v interface{}, out *[]string) {
	switch t := v.(type) {
	case map[string]interface{}:
		if isExplicitlyDisabled(t) {
			return
		}
		if img, ok := composeImageFromFields(t); ok {
			*out = append(*out, img)
		}
		if img := stringFromYAML(t["image"]); img != "" && strings.Contains(img, "/") {
			*out = append(*out, img)
		}
		for _, child := range t {
			collectImageFields(child, out)
		}
	case []interface{}:
		for _, child := range t {
			collectImageFields(child, out)
		}
	}
}

// composeImageFromFields builds a pullspec from common Helm image map fields.
// Charts like travelops set registry=quay.io and repository=kiali (org only) while the
// image name is hardcoded in the template (…/demo_travels_control). Those incomplete
// maps are skipped here; helm template discovery picks up the full refs.
func composeImageFromFields(t map[string]interface{}) (string, bool) {
	registry := strings.TrimSuffix(stringFromYAML(t["registry"]), "/")
	repo := strings.TrimPrefix(stringFromYAML(t["repository"]), "/")
	name := strings.TrimPrefix(stringFromYAML(t["name"]), "/")
	tag := stringFromYAML(t["tag"])
	if repo == "" && name == "" {
		return "", false
	}
	// registry + single-segment repository (no name) is an org/project prefix, not an image.
	if registry != "" && repo != "" && name == "" && !strings.Contains(repo, "/") {
		return "", false
	}
	img := repo
	if img == "" {
		img = name
		name = ""
	}
	if registry != "" {
		img = registry + "/" + img
	}
	if name != "" {
		img = img + "/" + name
	}
	if tag != "" {
		img = img + ":" + tag
	}
	return img, true
}

func isExplicitlyDisabled(m map[string]interface{}) bool {
	for _, key := range []string{"enabled", "enable"} {
		switch v := m[key].(type) {
		case bool:
			if !v {
				return true
			}
		case string:
			if strings.EqualFold(v, "false") {
				return true
			}
		}
	}
	return false
}

func stringFromYAML(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	default:
		return ""
	}
}

func imagesFromHelmTemplate(chartDir string) (images []string, warning string, err error) {
	if _, err := exec.LookPath("helm"); err != nil {
		return nil, "helm not found on PATH; skipping helm template discovery", nil
	}
	cmd := exec.Command("helm", "template", "airgap-discover", chartDir, "--include-crds")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", fmt.Errorf("%w: %s", err, truncate(string(out), 400))
	}
	return scrapeManifestImages(string(out)), "", nil
}

func scrapeManifestImages(manifest string) []string {
	var images []string
	for _, line := range strings.Split(manifest, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if !strings.HasPrefix(trimmed, "image:") {
			continue
		}
		val := strings.TrimSpace(strings.TrimPrefix(trimmed, "image:"))
		val = strings.Trim(val, `"'`)
		if val == "" || strings.HasPrefix(val, "{{") {
			continue
		}
		images = append(images, val)
	}
	return images
}

func filterRequiredImages(images map[string]struct{}, warnings []string) []string {
	out := make([]string, 0, len(images))
	for img := range images {
		if !isLikelyImage(img) {
			continue
		}
		if !isVPSourceImage(img) {
			continue
		}
		out = append(out, normalizeImagePullSpec(img))
	}
	return dropRedundantImagePrefixes(uniqueSorted(out))
}

func sliceToSet(in []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, s := range in {
		out[s] = struct{}{}
	}
	return out
}

func isVPSourceImage(img string) bool {
	img = strings.TrimSpace(img)
	if img == "" {
		return false
	}
	if strings.HasPrefix(img, "docker.io/") {
		return false
	}
	host, _, ok := splitRegistry(img)
	if !ok {
		return false
	}
	if isKubernetesServiceHost(host) {
		return false
	}
	switch host {
	case "registry.redhat.io",
		"registry.connect.redhat.com",
		"registry.access.redhat.com",
		"quay.io",
		"ghcr.io":
		return true
	default:
		if host == "docker.io" || host == "index.docker.io" || host == "registry-1.docker.io" {
			return false
		}
		return strings.Contains(host, ".")
	}
}

// isKubernetesServiceHost reports whether host is a cluster-local Kubernetes
// Service DNS name (e.g. image-registry.openshift-image-registry.svc:5000).
func isKubernetesServiceHost(host string) bool {
	h := host
	if i := strings.LastIndex(h, ":"); i >= 0 {
		h = h[:i]
	}
	return strings.HasSuffix(h, ".svc") || strings.Contains(h, ".svc.")
}

func splitRegistry(img string) (host, remainder string, ok bool) {
	parts := strings.SplitN(img, "/", 2)
	if len(parts) < 2 {
		return "", "", false
	}
	host = parts[0]
	if strings.Contains(host, ".") || strings.Contains(host, ":") {
		return host, parts[1], true
	}
	return "", "", false
}

func normalizeImagePullSpec(img string) string {
	return strings.TrimSpace(img)
}

func isLikelyImage(s string) bool {
	if s == "" || strings.Contains(s, "{{") || strings.Contains(s, "}}") {
		return false
	}
	if !strings.Contains(s, "/") {
		return false
	}
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "/") {
		return false
	}
	return hasCompleteImagePath(s)
}

// hasCompleteImagePath rejects incomplete specs like quay.io/kiali:v1 where values
// only supplied an org and the image name lives in the Helm template. Multi-tenant
// hosts need registry/org/name; other hosts may use a single path segment.
func hasCompleteImagePath(img string) bool {
	host, remainder, ok := splitRegistry(img)
	path := img
	if ok {
		path = remainder
	}
	path = stripImageTagOrDigest(path)
	n := 0
	for _, seg := range strings.Split(path, "/") {
		if seg != "" {
			n++
		}
	}
	if n < 1 {
		return false
	}
	if !ok {
		return n >= 2
	}
	if n >= 2 {
		return true
	}
	return !isMultiTenantRegistryHost(host)
}

func isMultiTenantRegistryHost(host string) bool {
	h := host
	if i := strings.LastIndex(h, ":"); i >= 0 && !strings.Contains(h[i+1:], "]") {
		h = h[:i]
	}
	switch h {
	case "quay.io",
		"ghcr.io",
		"docker.io",
		"index.docker.io",
		"registry-1.docker.io",
		"registry.redhat.io",
		"registry.connect.redhat.com",
		"registry.access.redhat.com",
		"gcr.io",
		"public.ecr.aws":
		return true
	default:
		return false
	}
}

func stripImageTagOrDigest(path string) string {
	if i := strings.Index(path, "@"); i >= 0 {
		path = path[:i]
	}
	if i := strings.LastIndex(path, ":"); i >= 0 {
		path = path[:i]
	}
	return path
}

func imageRepoPath(img string) string {
	img = strings.TrimSpace(img)
	host, rem, ok := splitRegistry(img)
	if !ok {
		return stripImageTagOrDigest(img)
	}
	return host + "/" + stripImageTagOrDigest(rem)
}

// dropRedundantImagePrefixes removes pullspecs whose repository path is a strict
// prefix of another collected image (e.g. quay.io/kiali when quay.io/kiali/app exists).
func dropRedundantImagePrefixes(images []string) []string {
	if len(images) < 2 {
		return images
	}
	paths := make([]string, len(images))
	for i, img := range images {
		paths[i] = imageRepoPath(img)
	}
	keep := make([]bool, len(images))
	for i := range images {
		keep[i] = true
		for j := range images {
			if i == j {
				continue
			}
			if paths[i] != paths[j] && strings.HasPrefix(paths[j], paths[i]+"/") {
				keep[i] = false
				break
			}
		}
	}
	out := make([]string, 0, len(images))
	for i, img := range images {
		if keep[i] {
			out = append(out, img)
		}
	}
	return out
}

func uniqueSorted(in []string) []string {
	set := map[string]struct{}{}
	for _, s := range in {
		if isLikelyImage(s) && isVPSourceImage(s) {
			set[normalizeImagePullSpec(s)] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
