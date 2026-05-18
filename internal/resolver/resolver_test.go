package resolver_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/config"
	"github.com/duyanh-y4n/dpod-seed/internal/fetch"
	"github.com/duyanh-y4n/dpod-seed/internal/resolver"
)

// mockFetcher records calls and returns pre-configured responses.
type mockFetcher struct {
	responses map[string][]fetch.File // key: "repo@sha:path"
	errors    map[string]error
}

func (m *mockFetcher) Fetch(repo, sha, path string) ([]fetch.File, error) {
	key := fmt.Sprintf("%s@%s:%s", repo, sha, path)
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	if files, ok := m.responses[key]; ok {
		return files, nil
	}
	return nil, fmt.Errorf("unexpected Fetch(%q, %q, %q)", repo, sha, path)
}

var repos = resolver.RepoConfig{
	DistroRepo:       "github.com/org/distros",
	DevcontainerRepo: "github.com/org/devcontainer",
	PackagesRepo:     "github.com/org/packages",
}

func TestResolve_singlePackage(t *testing.T) {
	f := &mockFetcher{
		responses: map[string][]fetch.File{
			"github.com/org/distros@v1.0.0:distros/myos": {
				{Path: "distro.yaml", Content: []byte(`
devcontainer: arch-base@v2.0.0
packages:
  - shell-zsh@v1.3.0
`)},
			},
			"github.com/org/devcontainer@v2.0.0:profiles/arch-base": {
				{Path: ".devcontainer/devcontainer.json", Content: []byte(`{"image":"base"}`)},
			},
			"github.com/org/packages@v1.3.0:packages/shell-zsh": {
				{Path: "dotfiles/.zshrc", Content: []byte("export ZSH=$HOME/.oh-my-zsh")},
			},
		},
	}

	r := resolver.NewResolver(f, repos)
	entries, err := r.Resolve("myos", "v1.0.0", config.Overrides{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries count = %d, want 2", len(entries))
	}

	// devcontainer profile file: path preserved as-is
	dcEntry := findEntry(entries, ".devcontainer/devcontainer.json")
	if dcEntry == nil {
		t.Fatalf("missing .devcontainer/devcontainer.json")
	}
	if dcEntry.SrcRepo != "github.com/org/devcontainer" {
		t.Errorf("SrcRepo = %q, want devcontainer repo", dcEntry.SrcRepo)
	}
	if dcEntry.SHA != "v2.0.0" {
		t.Errorf("SHA = %q, want v2.0.0", dcEntry.SHA)
	}

	// package bundle: dotfiles/.zshrc → .devcontainer/.zshrc
	pkgEntry := findEntry(entries, ".devcontainer/.zshrc")
	if pkgEntry == nil {
		t.Fatalf("missing .devcontainer/.zshrc")
	}
	if pkgEntry.SrcRepo != "github.com/org/packages" {
		t.Errorf("SrcRepo = %q, want packages repo", pkgEntry.SrcRepo)
	}
	if pkgEntry.SHA != "v1.3.0" {
		t.Errorf("SHA = %q, want v1.3.0", pkgEntry.SHA)
	}
	if string(pkgEntry.Content) != "export ZSH=$HOME/.oh-my-zsh" {
		t.Errorf("content = %q", string(pkgEntry.Content))
	}
}

func TestResolve_multiplePackages(t *testing.T) {
	f := &mockFetcher{
		responses: map[string][]fetch.File{
			"github.com/org/distros@v1.0.0:distros/full": {
				{Path: "distro.yaml", Content: []byte(`
devcontainer: arch-base@v2.0.0
packages:
  - shell-zsh@v1.3.0
  - k8s-tools@v1.1.0
`)},
			},
			"github.com/org/devcontainer@v2.0.0:profiles/arch-base": {
				{Path: ".devcontainer/devcontainer.json", Content: []byte(`{}`)},
			},
			"github.com/org/packages@v1.3.0:packages/shell-zsh": {
				{Path: "dotfiles/.zshrc", Content: []byte("zsh config")},
			},
			"github.com/org/packages@v1.1.0:packages/k8s-tools": {
				{Path: "dotfiles/.kube/config", Content: []byte("kube config")},
			},
		},
	}

	r := resolver.NewResolver(f, repos)
	entries, err := r.Resolve("full", "v1.0.0", config.Overrides{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// 1 devcontainer file + 1 zsh + 1 kube = 3
	if len(entries) != 3 {
		t.Fatalf("entries count = %d, want 3", len(entries))
	}
	// k8s-tools uses its own SHA v1.1.0, not the distro tag v1.0.0
	kubeEntry := findEntry(entries, ".devcontainer/.kube/config")
	if kubeEntry == nil {
		t.Fatalf("missing .devcontainer/.kube/config")
	}
	if kubeEntry.SHA != "v1.1.0" {
		t.Errorf("SHA = %q, want v1.1.0", kubeEntry.SHA)
	}
}

func TestResolve_manifestYamlExcluded(t *testing.T) {
	f := &mockFetcher{
		responses: map[string][]fetch.File{
			"github.com/org/distros@v1.0.0:distros/myos": {
				{Path: "distro.yaml", Content: []byte("devcontainer: arch-base@v2.0.0\npackages:\n  - shell-zsh@v1.3.0\n")},
			},
			"github.com/org/devcontainer@v2.0.0:profiles/arch-base": {},
			"github.com/org/packages@v1.3.0:packages/shell-zsh": {
				{Path: "manifest.yaml", Content: []byte("files: []")},
				{Path: "dotfiles/.zshrc", Content: []byte("zsh")},
			},
		},
	}

	r := resolver.NewResolver(f, repos)
	entries, err := r.Resolve("myos", "v1.0.0", config.Overrides{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	for _, e := range entries {
		if e.DestPath == "manifest.yaml" {
			t.Error("manifest.yaml should be excluded from output")
		}
	}
	if len(entries) != 1 {
		t.Fatalf("entries count = %d, want 1 (.zshrc only)", len(entries))
	}
}

func TestResolve_overridesAdd(t *testing.T) {
	f := &mockFetcher{
		responses: map[string][]fetch.File{
			"github.com/org/distros@v1.0.0:distros/myos": {
				{Path: "distro.yaml", Content: []byte("devcontainer: arch-base@v2.0.0\npackages: []\n")},
			},
			"github.com/org/devcontainer@v2.0.0:profiles/arch-base": {},
			// override neovim fetched at the distro tag v1.0.0
			"github.com/org/packages@v1.0.0:packages/neovim": {
				{Path: "dotfiles/.config/nvim/init.lua", Content: []byte("neovim config")},
			},
		},
	}

	r := resolver.NewResolver(f, repos)
	overrides := config.Overrides{
		Packages: config.PackageOverrides{Add: []string{"neovim"}},
	}
	entries, err := r.Resolve("myos", "v1.0.0", overrides)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	nvimEntry := findEntry(entries, ".devcontainer/.config/nvim/init.lua")
	if nvimEntry == nil {
		t.Fatalf("missing neovim entry")
	}
	if nvimEntry.SHA != "v1.0.0" {
		t.Errorf("SHA = %q, want v1.0.0 (distro tag)", nvimEntry.SHA)
	}
}

func TestResolve_overridesRemove(t *testing.T) {
	f := &mockFetcher{
		responses: map[string][]fetch.File{
			"github.com/org/distros@v1.0.0:distros/myos": {
				{Path: "distro.yaml", Content: []byte("devcontainer: arch-base@v2.0.0\npackages:\n  - shell-zsh@v1.3.0\n  - vscode@v1.0.0\n")},
			},
			"github.com/org/devcontainer@v2.0.0:profiles/arch-base": {},
			"github.com/org/packages@v1.3.0:packages/shell-zsh": {
				{Path: "dotfiles/.zshrc", Content: []byte("zsh")},
			},
			// vscode should never be fetched
		},
	}

	r := resolver.NewResolver(f, repos)
	overrides := config.Overrides{
		Packages: config.PackageOverrides{Remove: []string{"vscode"}},
	}
	entries, err := r.Resolve("myos", "v1.0.0", overrides)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	for _, e := range entries {
		if strings.Contains(e.DestPath, "vscode") {
			t.Errorf("vscode entry should be removed, got %q", e.DestPath)
		}
	}
	if len(entries) != 1 {
		t.Fatalf("entries count = %d, want 1 (zsh only)", len(entries))
	}
}

func TestResolve_manifestYamlOverridesDest(t *testing.T) {
	f := &mockFetcher{
		responses: map[string][]fetch.File{
			"github.com/org/distros@v1.0.0:distros/myos": {
				{Path: "distro.yaml", Content: []byte("devcontainer: arch-base@v2.0.0\npackages:\n  - shell-zsh@v1.3.0\n")},
			},
			"github.com/org/devcontainer@v2.0.0:profiles/arch-base": {},
			"github.com/org/packages@v1.3.0:packages/shell-zsh": {
				{Path: "manifest.yaml", Content: []byte("files:\n  - src: dotfiles/.zshrc\n    dest: home/.zshrc\n")},
				{Path: "dotfiles/.zshrc", Content: []byte("zsh config")},
			},
		},
	}

	r := resolver.NewResolver(f, repos)
	entries, err := r.Resolve("myos", "v1.0.0", config.Overrides{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// The manifest override should change dotfiles/.zshrc → home/.zshrc
	// instead of the default .devcontainer/.zshrc
	e := findEntry(entries, "home/.zshrc")
	if e == nil {
		t.Fatalf("expected entry at home/.zshrc (manifest override), got %v", entryPaths(entries))
	}
	// and the default-convention path should NOT appear
	if findEntry(entries, ".devcontainer/.zshrc") != nil {
		t.Error("default-convention path should be overridden by manifest.yaml")
	}
}

func entryPaths(entries []resolver.ManifestEntry) []string {
	paths := make([]string, len(entries))
	for i, e := range entries {
		paths[i] = e.DestPath
	}
	return paths
}

func TestResolve_missingDistroYAML(t *testing.T) {
	f := &mockFetcher{
		responses: map[string][]fetch.File{
			// fetcher succeeds but returns no distro.yaml
			"github.com/org/distros@v1.0.0:distros/myos": {
				{Path: "README.md", Content: []byte("# myos")},
			},
		},
	}
	r := resolver.NewResolver(f, repos)
	_, err := r.Resolve("myos", "v1.0.0", config.Overrides{})
	if err == nil {
		t.Fatal("expected error when distro.yaml is absent from fetched files")
	}
	if !strings.Contains(err.Error(), "distro.yaml") {
		t.Errorf("error should mention distro.yaml, got: %v", err)
	}
}

func TestResolve_unknownDistro(t *testing.T) {
	f := &mockFetcher{
		errors: map[string]error{
			"github.com/org/distros@v1.0.0:distros/badname": fmt.Errorf("not found"),
		},
	}

	r := resolver.NewResolver(f, repos)
	_, err := r.Resolve("badname", "v1.0.0", config.Overrides{})
	if err == nil {
		t.Fatal("expected error for unknown distro")
	}
	if !strings.Contains(err.Error(), "badname") {
		t.Errorf("error should mention distro name, got: %v", err)
	}
}

func TestResolve_nonExistentTag(t *testing.T) {
	f := &mockFetcher{
		errors: map[string]error{
			"github.com/org/distros@v9.9.9:distros/myos": fmt.Errorf("not found"),
		},
	}

	r := resolver.NewResolver(f, repos)
	_, err := r.Resolve("myos", "v9.9.9", config.Overrides{})
	if err == nil {
		t.Fatal("expected error for non-existent tag")
	}
	if !strings.Contains(err.Error(), "v9.9.9") {
		t.Errorf("error should mention tag, got: %v", err)
	}
}

func findEntry(entries []resolver.ManifestEntry, destPath string) *resolver.ManifestEntry {
	for i := range entries {
		if entries[i].DestPath == destPath {
			return &entries[i]
		}
	}
	return nil
}
