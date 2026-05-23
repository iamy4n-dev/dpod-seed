package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/iamy4n-dev/dpod-seed/internal/fetch"
	"github.com/iamy4n-dev/dpod-seed/internal/registry"
)

// Package is a distro package entry with its name and pinned version.
type Package struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// DistroRecord is one entry in the generated registry-data.json.
type DistroRecord struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	LatestTag    string    `json:"latestTag"`
	Status       string    `json:"status"`
	ChangelogURL string    `json:"changelogUrl,omitempty"`
	Packages     []Package `json:"packages"`
}

// Output is the top-level structure of registry-data.json.
type Output struct {
	Distros []DistroRecord `json:"distros"`
}

type distroYAML struct {
	Packages []string `yaml:"packages"`
}

// Generate fetches each distro manifest and writes registry-data.json to w.
func Generate(reg registry.Client, f fetch.Fetcher, distroRepo string, w io.Writer) error {
	entries, err := reg.List()
	if err != nil {
		return fmt.Errorf("list registry: %w", err)
	}

	var records []DistroRecord
	for _, e := range entries {
		path := "distros/" + e.Name + "/distro.yaml"
		files, err := f.Fetch(distroRepo, e.LatestTag, path)
		if err != nil {
			return fmt.Errorf("fetch distro %s at %s: %w", e.Name, e.LatestTag, err)
		}
		var content []byte
		for _, file := range files {
			if strings.HasSuffix(file.Path, "distro.yaml") || file.Path == "distro.yaml" {
				content = file.Content
				break
			}
		}
		if content == nil {
			return fmt.Errorf("distro %s: distro.yaml not found in fetch result", e.Name)
		}
		var d distroYAML
		if err := yaml.Unmarshal(content, &d); err != nil {
			return fmt.Errorf("parse distro %s: %w", e.Name, err)
		}
		pkgs := make([]Package, 0, len(d.Packages))
		for _, p := range d.Packages {
			if i := strings.LastIndex(p, "@"); i >= 0 {
				pkgs = append(pkgs, Package{Name: p[:i], Version: p[i+1:]})
			} else {
				pkgs = append(pkgs, Package{Name: p})
			}
		}
		records = append(records, DistroRecord{
			Name:         e.Name,
			Description:  e.Description,
			LatestTag:    e.LatestTag,
			Status:       e.Status,
			ChangelogURL: e.ChangelogURL,
			Packages:     pkgs,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(Output{Distros: records})
}
