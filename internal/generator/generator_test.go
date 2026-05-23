package generator_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/iamy4n-dev/dpod-seed/internal/fetch"
	"github.com/iamy4n-dev/dpod-seed/internal/generator"
	"github.com/iamy4n-dev/dpod-seed/internal/registry"
)

// stubRegistryClient returns a fixed list of distro entries.
type stubRegistryClient struct {
	entries []registry.DistroEntry
	err     error
}

func (s *stubRegistryClient) List() ([]registry.DistroEntry, error) {
	return s.entries, s.err
}

// stubFetcher returns fixed file contents per "repo@sha/path" key.
type stubFetcher struct {
	files map[string][]fetch.File
	err   error
}

func (s *stubFetcher) Fetch(repo, sha, path string) ([]fetch.File, error) {
	if s.err != nil {
		return nil, s.err
	}
	key := repo + "@" + sha + "/" + path
	if files, ok := s.files[key]; ok {
		return files, nil
	}
	return nil, fmt.Errorf("no stub for %s", key)
}

const devopsDistroYAML = `name: devops-k8s
description: Kubernetes development environment
devcontainer: arch-base@v2.0.0
packages:
  - shell-zsh@v1.3.0
  - k8s-tools@v1.1.0
`

const devopsReadme = `# devops-k8s

A Kubernetes development distro with k8s-tools and zsh.
`

// --- Tracer bullet ---

func TestGenerate_singleDistro_validJSON(t *testing.T) {
	reg := &stubRegistryClient{entries: []registry.DistroEntry{
		{Name: "devops-k8s", Description: "Kubernetes development environment", LatestTag: "v0.2.0", Status: "stable", ChangelogURL: "https://example.com/changelog"},
	}}
	f := &stubFetcher{files: map[string][]fetch.File{
		"github.com/iamy4n-dev/distros@v0.2.0/distros/devops-k8s/distro.yaml": {
			{Path: "distro.yaml", Content: []byte(devopsDistroYAML)},
		},
	}}

	var buf strings.Builder
	if err := generator.Generate(reg, f, "github.com/iamy4n-dev/distros", &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var out generator.Output
	if err := json.Unmarshal([]byte(buf.String()), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.Distros) != 1 {
		t.Fatalf("expected 1 distro, got %d", len(out.Distros))
	}
	d := out.Distros[0]
	if d.Name != "devops-k8s" {
		t.Errorf("name = %q, want devops-k8s", d.Name)
	}
	if d.LatestTag != "v0.2.0" {
		t.Errorf("latestTag = %q, want v0.2.0", d.LatestTag)
	}
	if d.Status != "stable" {
		t.Errorf("status = %q, want stable", d.Status)
	}
}

// --- Package versions preserved as structured fields ---

func TestGenerate_packageVersionsPreserved(t *testing.T) {
	reg := &stubRegistryClient{entries: []registry.DistroEntry{
		{Name: "devops-k8s", Description: "K8s env", LatestTag: "v0.2.0", Status: "stable"},
	}}
	f := &stubFetcher{files: map[string][]fetch.File{
		"github.com/iamy4n-dev/distros@v0.2.0/distros/devops-k8s/distro.yaml": {
			{Path: "distro.yaml", Content: []byte(devopsDistroYAML)},
		},
	}}

	var buf strings.Builder
	if err := generator.Generate(reg, f, "github.com/iamy4n-dev/distros", &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var out generator.Output
	json.Unmarshal([]byte(buf.String()), &out)
	pkgs := out.Distros[0].Packages
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "shell-zsh" {
		t.Errorf("pkgs[0].Name = %q, want shell-zsh", pkgs[0].Name)
	}
	if pkgs[0].Version != "v1.3.0" {
		t.Errorf("pkgs[0].Version = %q, want v1.3.0", pkgs[0].Version)
	}
	if pkgs[1].Name != "k8s-tools" {
		t.Errorf("pkgs[1].Name = %q, want k8s-tools", pkgs[1].Name)
	}
	if pkgs[1].Version != "v1.1.0" {
		t.Errorf("pkgs[1].Version = %q, want v1.1.0", pkgs[1].Version)
	}
}

// --- Multiple distros ---

func TestGenerate_multipleDistros(t *testing.T) {
	reg := &stubRegistryClient{entries: []registry.DistroEntry{
		{Name: "distro-a", Description: "A", LatestTag: "v1.0.0", Status: "stable"},
		{Name: "distro-b", Description: "B", LatestTag: "v2.0.0", Status: "experimental"},
	}}
	const simpleYAML = "name: x\ndescription: x\ndevcontainer: base@v1.0.0\npackages: []\n"
	f := &stubFetcher{files: map[string][]fetch.File{
		"github.com/iamy4n-dev/distros@v1.0.0/distros/distro-a/distro.yaml": {
			{Path: "distro.yaml", Content: []byte(simpleYAML)},
		},
		"github.com/iamy4n-dev/distros@v2.0.0/distros/distro-b/distro.yaml": {
			{Path: "distro.yaml", Content: []byte(simpleYAML)},
		},
	}}

	var buf strings.Builder
	if err := generator.Generate(reg, f, "github.com/iamy4n-dev/distros", &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var out generator.Output
	json.Unmarshal([]byte(buf.String()), &out)
	if len(out.Distros) != 2 {
		t.Errorf("expected 2 distros, got %d", len(out.Distros))
	}
}

// --- Fetch error ---

func TestGenerate_fetchError_returnsError(t *testing.T) {
	reg := &stubRegistryClient{entries: []registry.DistroEntry{
		{Name: "broken", Description: "broken", LatestTag: "v1.0.0", Status: "stable"},
	}}
	f := &stubFetcher{err: fmt.Errorf("network error")}

	var buf strings.Builder
	err := generator.Generate(reg, f, "github.com/iamy4n-dev/distros", &buf)
	if err == nil {
		t.Error("expected error from fetch failure")
	}
}

// --- Registry list error ---

func TestGenerate_registryError_returnsError(t *testing.T) {
	reg := &stubRegistryClient{err: fmt.Errorf("registry unavailable")}

	var buf strings.Builder
	err := generator.Generate(reg, &stubFetcher{}, "github.com/iamy4n-dev/distros", &buf)
	if err == nil {
		t.Error("expected error from registry failure")
	}
}

// --- README content ---

func TestGenerate_sourceURL_isGitHubTreeURL(t *testing.T) {
	reg := &stubRegistryClient{entries: []registry.DistroEntry{
		{Name: "devops-k8s", Description: "K8s env", LatestTag: "v0.2.0", Status: "stable"},
	}}
	f := &stubFetcher{files: map[string][]fetch.File{
		"github.com/iamy4n-dev/distros@v0.2.0/distros/devops-k8s/distro.yaml": {
			{Path: "distro.yaml", Content: []byte(devopsDistroYAML)},
		},
	}}

	var buf strings.Builder
	if err := generator.Generate(reg, f, "github.com/iamy4n-dev/distros", &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var out generator.Output
	json.Unmarshal([]byte(buf.String()), &out)
	want := "https://github.com/iamy4n-dev/distros/tree/v0.2.0/distros/devops-k8s"
	if out.Distros[0].SourceURL != want {
		t.Errorf("sourceUrl = %q, want %q", out.Distros[0].SourceURL, want)
	}
}

func TestGenerate_withoutReadme_noErrorAndEmptyField(t *testing.T) {
	reg := &stubRegistryClient{entries: []registry.DistroEntry{
		{Name: "devops-k8s", Description: "K8s env", LatestTag: "v0.2.0", Status: "stable"},
	}}
	f := &stubFetcher{files: map[string][]fetch.File{
		"github.com/iamy4n-dev/distros@v0.2.0/distros/devops-k8s/distro.yaml": {
			{Path: "distro.yaml", Content: []byte(devopsDistroYAML)},
		},
		// no README.md entry
	}}

	var buf strings.Builder
	if err := generator.Generate(reg, f, "github.com/iamy4n-dev/distros", &buf); err != nil {
		t.Fatalf("Generate returned error with no README: %v", err)
	}

	var out generator.Output
	json.Unmarshal([]byte(buf.String()), &out)
	if out.Distros[0].Readme != "" {
		t.Errorf("expected empty readme, got %q", out.Distros[0].Readme)
	}
}

func TestGenerate_withReadme_populatesReadmeField(t *testing.T) {
	reg := &stubRegistryClient{entries: []registry.DistroEntry{
		{Name: "devops-k8s", Description: "K8s env", LatestTag: "v0.2.0", Status: "stable"},
	}}
	f := &stubFetcher{files: map[string][]fetch.File{
		"github.com/iamy4n-dev/distros@v0.2.0/distros/devops-k8s/distro.yaml": {
			{Path: "distro.yaml", Content: []byte(devopsDistroYAML)},
		},
		"github.com/iamy4n-dev/distros@v0.2.0/distros/devops-k8s/README.md": {
			{Path: "README.md", Content: []byte(devopsReadme)},
		},
	}}

	var buf strings.Builder
	if err := generator.Generate(reg, f, "github.com/iamy4n-dev/distros", &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var out generator.Output
	json.Unmarshal([]byte(buf.String()), &out)
	d := out.Distros[0]
	if d.Readme == "" {
		t.Error("expected readme to be populated")
	}
	if !strings.Contains(d.Readme, "devops-k8s") {
		t.Errorf("readme = %q, expected to contain 'devops-k8s'", d.Readme)
	}
}
