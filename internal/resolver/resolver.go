package resolver

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/iamy4n-dev/dpod-seed/internal/config"
	"github.com/iamy4n-dev/dpod-seed/internal/fetch"
)

type RepoConfig struct {
	DistroRepo       string
	DevcontainerRepo string
	PackagesRepo     string
}

type ManifestEntry struct {
	DestPath string
	SrcRepo  string
	SHA      string
	Content  []byte
}

type Resolver interface {
	Resolve(distroName, tag string, overrides config.Overrides) ([]ManifestEntry, error)
}

type resolverImpl struct {
	fetcher fetch.Fetcher
	repos   RepoConfig
}

func NewResolver(fetcher fetch.Fetcher, repos RepoConfig) Resolver {
	return &resolverImpl{fetcher: fetcher, repos: repos}
}

type distroManifest struct {
	Devcontainer string   `yaml:"devcontainer"`
	Packages     []string `yaml:"packages"`
}

func (r *resolverImpl) Resolve(distroName, tag string, overrides config.Overrides) ([]ManifestEntry, error) {
	dm, err := r.fetchDistroManifest(distroName, tag)
	if err != nil {
		return nil, fmt.Errorf("resolve distro %q@%s: %w", distroName, tag, err)
	}

	var entries []ManifestEntry

	// Devcontainer profile
	dcEntries, err := r.resolveDevcontainer(dm.Devcontainer)
	if err != nil {
		return nil, err
	}
	entries = append(entries, dcEntries...)

	// Package bundles from distro.yaml, minus removes
	removeSet := toSet(overrides.Packages.Remove)
	for _, pin := range dm.Packages {
		name, sha := splitPin(pin)
		if removeSet[name] {
			continue
		}
		pkgEntries, err := r.resolvePackage(name, sha)
		if err != nil {
			return nil, err
		}
		entries = append(entries, pkgEntries...)
	}

	// overrides.add — each entry must carry an explicit version pin (name@tag).
	for _, pin := range overrides.Packages.Add {
		name, sha := splitPin(pin)
		if sha == "" {
			return nil, fmt.Errorf("overrides.add entry %q must include a version pin, e.g. %s@<tag>", pin, pin)
		}
		pkgEntries, err := r.resolvePackage(name, sha)
		if err != nil {
			return nil, err
		}
		entries = append(entries, pkgEntries...)
	}

	return entries, nil
}

func (r *resolverImpl) fetchDistroManifest(distroName, tag string) (*distroManifest, error) {
	files, err := r.fetcher.Fetch(r.repos.DistroRepo, tag, "distros/"+distroName)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.Path == "distro.yaml" {
			var dm distroManifest
			if err := yaml.Unmarshal(f.Content, &dm); err != nil {
				return nil, fmt.Errorf("parse distro.yaml: %w", err)
			}
			return &dm, nil
		}
	}
	return nil, fmt.Errorf("distro.yaml not found in distros/%s", distroName)
}

func (r *resolverImpl) resolveDevcontainer(pin string) ([]ManifestEntry, error) {
	name, sha := splitPin(pin)
	files, err := r.fetcher.Fetch(r.repos.DevcontainerRepo, sha, "profiles/"+name)
	if err != nil {
		return nil, fmt.Errorf("fetch devcontainer profile %q@%s: %w", name, sha, err)
	}
	entries := make([]ManifestEntry, 0, len(files))
	for _, f := range files {
		entries = append(entries, ManifestEntry{
			DestPath: f.Path,
			SrcRepo:  r.repos.DevcontainerRepo,
			SHA:      sha,
			Content:  f.Content,
		})
	}
	return entries, nil
}

type packageManifest struct {
	Files []struct {
		Src  string `yaml:"src"`
		Dest string `yaml:"dest"`
	} `yaml:"files"`
}

func (r *resolverImpl) resolvePackage(name, sha string) ([]ManifestEntry, error) {
	files, err := r.fetcher.Fetch(r.repos.PackagesRepo, sha, "packages/"+name)
	if err != nil {
		return nil, fmt.Errorf("fetch package %q@%s: %w", name, sha, err)
	}

	// Build override map from manifest.yaml if present
	overrides := map[string]string{}
	for _, f := range files {
		if f.Path == "manifest.yaml" {
			var pm packageManifest
			if err := yaml.Unmarshal(f.Content, &pm); err == nil {
				for _, rule := range pm.Files {
					overrides[rule.Src] = rule.Dest
				}
			}
		}
	}

	entries := make([]ManifestEntry, 0, len(files))
	for _, f := range files {
		if f.Path == "manifest.yaml" {
			continue
		}
		var dest string
		if d, ok := overrides[f.Path]; ok {
			dest = d
		} else {
			dest = applyPlacement(f.Path)
		}
		entries = append(entries, ManifestEntry{
			DestPath: dest,
			SrcRepo:  r.repos.PackagesRepo,
			SHA:      sha,
			Content:  f.Content,
		})
	}
	return entries, nil
}

// applyPlacement maps dotfiles/<path> → .devcontainer/<path>; other paths pass through.
func applyPlacement(srcPath string) string {
	if rest, ok := strings.CutPrefix(srcPath, "dotfiles/"); ok {
		return ".devcontainer/" + rest
	}
	return srcPath
}

// splitPin splits "name@tag" into (name, tag). If no "@", tag is empty string.
func splitPin(pin string) (name, tag string) {
	if i := strings.LastIndex(pin, "@"); i >= 0 {
		return pin[:i], pin[i+1:]
	}
	return pin, ""
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
