package pattern

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/validatedpatterns/patternizer/internal/embedded"
)

type metadataTemplateData struct {
	Name        string
	DisplayName string
	RepoURL     string
	IssuesURL   string
	Org         string
}

// GeneratePatternMetadata creates a pattern-metadata.yaml file in the repo root.
// It skips generation if the file already exists.
func GeneratePatternMetadata(patternName, repoRoot string) error {
	metadataPath := filepath.Join(repoRoot, "pattern-metadata.yaml")
	if _, err := os.Stat(metadataPath); err == nil {
		return nil
	}

	org, repoURL := detectGitRemote(patternName, repoRoot)

	data := metadataTemplateData{
		Name:        patternName,
		DisplayName: toDisplayName(patternName),
		RepoURL:     repoURL,
		IssuesURL:   repoURL + "/issues",
		Org:         org,
	}

	tmplBytes, err := embedded.Resources.ReadFile("resources/pattern-metadata.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("reading metadata template: %w", err)
	}

	tmpl, err := template.New("metadata").Parse(string(tmplBytes))
	if err != nil {
		return fmt.Errorf("parsing metadata template: %w", err)
	}

	f, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", metadataPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing metadata template: %w", err)
	}

	return os.Chmod(metadataPath, 0o644)
}

func detectGitRemote(patternName, repoRoot string) (org, repoURL string) {
	fallbackOrg := "CHANGEME"
	fallbackURL := "https://github.com/" + fallbackOrg + "/" + patternName

	out, err := exec.Command("git", "-C", repoRoot, "remote", "get-url", "origin").Output()
	if err != nil {
		return fallbackOrg, fallbackURL
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return fallbackOrg, fallbackURL
	}

	return parseGitRemoteURL(raw, fallbackOrg, fallbackURL)
}

func parseGitRemoteURL(raw, fallbackOrg, fallbackURL string) (org, repoURL string) {
	if strings.HasPrefix(raw, "git@") {
		raw = strings.TrimPrefix(raw, "git@")
		raw = strings.Replace(raw, ":", "/", 1)
		raw = "https://" + raw
	}

	raw = strings.TrimSuffix(raw, ".git")

	parsed, err := url.Parse(raw)
	if err != nil {
		return fallbackOrg, fallbackURL
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return fallbackOrg, fallbackURL
	}

	org = parts[0]
	httpsURL := fmt.Sprintf("https://%s/%s/%s", parsed.Host, parts[0], parts[1])
	return org, httpsURL
}

func toDisplayName(name string) string {
	words := strings.Split(name, "-")
	for i, w := range words {
		if w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
