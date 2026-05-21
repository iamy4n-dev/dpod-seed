package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var pinRE = regexp.MustCompile(`^[a-zA-Z0-9_/.-]+@v\d+\.\d+\.\d+$`)

// Run auto-detects whether root is a distros, packages, or profiles repo and validates all entries.
func Run(root string) []string {
	distrosDir := filepath.Join(root, "distros")
	packagesDir := filepath.Join(root, "packages")
	profilesDir := filepath.Join(root, "profiles")

	hasDistros := dirExists(distrosDir)
	hasPackages := dirExists(packagesDir)
	hasProfiles := dirExists(profilesDir)

	if !hasDistros && !hasPackages && !hasProfiles {
		return []string{fmt.Sprintf("%s: no distros/, packages/, or profiles/ directory found", root)}
	}

	var errs []string
	if hasDistros {
		errs = append(errs, walkDistros(distrosDir)...)
	}
	if hasPackages {
		errs = append(errs, walkAndValidate(packagesDir, PackageDir)...)
	}
	if hasProfiles {
		errs = append(errs, walkAndValidate(profilesDir, ProfileDir)...)
	}
	return errs
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// walkDistros handles both flat (distros/{name}) and org-namespaced (distros/{org}/{name}) layouts.
// A directory is a leaf distro if it contains distro.yaml; otherwise it is treated as an org namespace.
func walkDistros(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read directory: %v", root, err)}
	}
	var errs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "distro.yaml")); err == nil {
			errs = append(errs, DistroDir(dir)...)
		} else {
			errs = append(errs, walkAndValidate(dir, DistroDir)...)
		}
	}
	return errs
}

func walkAndValidate(root string, validate func(string) []string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read directory: %v", root, err)}
	}
	var errs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		errs = append(errs, validate(filepath.Join(root, e.Name()))...)
	}
	return errs
}

type profileFrontmatter struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"display_name"`
	Description string   `yaml:"description"`
	Status      string   `yaml:"status"`
	Tags        []string `yaml:"tags"`
}

// ProfileDir validates a single profile directory (devcontainer.json + Dockerfile + README.md).
func ProfileDir(dir string) []string {
	var errs []string
	errs = append(errs, checkFileExists(filepath.Join(dir, "devcontainer.json"))...)
	errs = append(errs, checkFileExists(filepath.Join(dir, "Dockerfile"))...)
	errs = append(errs, validateProfileReadme(filepath.Join(dir, "README.md"))...)
	return errs
}

func checkFileExists(path string) []string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: file not found", path)}
	}
	return nil
}

func validateProfileReadme(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read file: %v", path, err)}
	}
	block, err := parseFrontmatterRaw(data)
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", path, err)}
	}
	var fm profileFrontmatter
	if err := yaml.Unmarshal([]byte(block), &fm); err != nil {
		return []string{fmt.Sprintf("%s: invalid frontmatter YAML: %v", path, err)}
	}
	var errs []string
	if strings.TrimSpace(fm.Name) == "" {
		errs = append(errs, fmt.Sprintf("%s: frontmatter missing required field \"name\"", path))
	}
	if strings.TrimSpace(fm.DisplayName) == "" {
		errs = append(errs, fmt.Sprintf("%s: frontmatter missing required field \"display_name\"", path))
	}
	if strings.TrimSpace(fm.Description) == "" {
		errs = append(errs, fmt.Sprintf("%s: frontmatter missing required field \"description\"", path))
	}
	if fm.Status != "experimental" && fm.Status != "stable" {
		errs = append(errs, fmt.Sprintf("%s: frontmatter status %q must be \"experimental\" or \"stable\"", path, fm.Status))
	}
	if len(fm.Tags) == 0 {
		errs = append(errs, fmt.Sprintf("%s: frontmatter missing required field \"tags\" (must have at least one)", path))
	}
	return errs
}

type distroYAML struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Devcontainer string  `yaml:"devcontainer"`
	Packages    []string `yaml:"packages"`
}

type readmeFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Status      string `yaml:"status"`
}

// DistroDir validates a single distro directory (distro.yaml + README.md).
func DistroDir(dir string) []string {
	var errs []string
	errs = append(errs, validateDistroYAML(filepath.Join(dir, "distro.yaml"))...)
	errs = append(errs, validateReadme(filepath.Join(dir, "README.md"), "distro")...)
	return errs
}

// PackageDir validates a single package directory (manifest.yaml + README.md).
func PackageDir(dir string) []string {
	var errs []string
	errs = append(errs, validateManifestYAML(filepath.Join(dir, "manifest.yaml"))...)
	errs = append(errs, validateReadme(filepath.Join(dir, "README.md"), "package")...)
	return errs
}

func validateDistroYAML(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read file: %v", path, err)}
	}
	var d distroYAML
	if err := yaml.Unmarshal(data, &d); err != nil {
		return []string{fmt.Sprintf("%s: invalid YAML: %v", path, err)}
	}
	var errs []string
	if strings.TrimSpace(d.Name) == "" {
		errs = append(errs, fmt.Sprintf("%s: missing required field \"name\"", path))
	}
	if strings.TrimSpace(d.Description) == "" {
		errs = append(errs, fmt.Sprintf("%s: missing required field \"description\"", path))
	}
	if strings.TrimSpace(d.Devcontainer) == "" {
		errs = append(errs, fmt.Sprintf("%s: missing required field \"devcontainer\"", path))
	} else if !pinRE.MatchString(d.Devcontainer) {
		errs = append(errs, fmt.Sprintf("%s: devcontainer pin %q is not in name@vX.Y.Z format", path, d.Devcontainer))
	}
	for _, pkg := range d.Packages {
		if !pinRE.MatchString(pkg) {
			errs = append(errs, fmt.Sprintf("%s: package pin %q is not in name@vX.Y.Z format", path, pkg))
		}
	}
	return errs
}

func validateManifestYAML(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read file: %v", path, err)}
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return []string{fmt.Sprintf("%s: invalid YAML: %v", path, err)}
	}
	return nil
}

func validateReadme(path, kind string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read file: %v", path, err)}
	}
	fm, err := parseFrontmatter(data)
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", path, err)}
	}
	var errs []string
	if strings.TrimSpace(fm.Name) == "" {
		errs = append(errs, fmt.Sprintf("%s: frontmatter missing required field \"name\"", path))
	}
	if strings.TrimSpace(fm.Description) == "" {
		errs = append(errs, fmt.Sprintf("%s: frontmatter missing required field \"description\"", path))
	}
	if fm.Status != "experimental" && fm.Status != "stable" {
		errs = append(errs, fmt.Sprintf("%s: frontmatter status %q must be \"experimental\" or \"stable\"", path, fm.Status))
	}
	_ = kind
	return errs
}

func parseFrontmatterRaw(data []byte) (string, error) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return "", fmt.Errorf("README.md: missing YAML frontmatter (must start with ---)")
	}
	end := strings.Index(s[4:], "\n---\n")
	if end == -1 {
		return "", fmt.Errorf("README.md: unclosed YAML frontmatter")
	}
	return s[4 : 4+end], nil
}

func parseFrontmatter(data []byte) (readmeFrontmatter, error) {
	block, err := parseFrontmatterRaw(data)
	if err != nil {
		return readmeFrontmatter{}, err
	}
	var fm readmeFrontmatter
	if err := yaml.Unmarshal([]byte(block), &fm); err != nil {
		return readmeFrontmatter{}, fmt.Errorf("README.md: invalid frontmatter YAML: %v", err)
	}
	return fm, nil
}
