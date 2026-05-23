package validate_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/iamy4n-dev/dpod-seed/internal/validate"
)

type stubRefChecker struct {
	results map[string]bool
	errs    map[string]error
}

func (s *stubRefChecker) TagExists(repo, tag string) (bool, error) {
	key := repo + "@" + tag
	if err, ok := s.errs[key]; ok {
		return false, err
	}
	exists, ok := s.results[key]
	if !ok {
		return true, nil
	}
	return exists, nil
}

const devcontainerRepo = "github.com/iamy4n-dev/devcontainer"
const packagesRepo = "github.com/iamy4n-dev/packages"

func refOpts(checker validate.RefChecker) validate.Options {
	return validate.Options{
		DevcontainerRepo: devcontainerRepo,
		PackagesRepo:     packagesRepo,
		Checker:          checker,
	}
}

func TestRefCheck_devcontainerTagMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/distro.yaml", "name: my-distro\ndescription: A distro\ndevcontainer: arch-base@v2.0.0\npackages: []\n")
	writeFile(t, dir+"/README.md", validDistroReadme)

	checker := &stubRefChecker{
		results: map[string]bool{devcontainerRepo + "@v2.0.0": false},
	}

	errs := validate.DistroDirWithOptions(dir, refOpts(checker))
	found := false
	for _, e := range errs {
		if strings.Contains(e, "arch-base@v2.0.0") && strings.Contains(e, "not found") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error for missing devcontainer tag, got: %v", errs)
	}
}

func TestRefCheck_offlineSkipsRefChecks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/distro.yaml", "name: my-distro\ndescription: A distro\ndevcontainer: arch-base@v99.0.0\npackages:\n  - shell-zsh@v99.0.0\n")
	writeFile(t, dir+"/README.md", validDistroReadme)

	errs := validate.DistroDirWithOptions(dir, validate.Options{Checker: nil})
	if len(errs) != 0 {
		t.Errorf("expected no errors in offline mode, got: %v", errs)
	}
}

func TestRefCheck_packageTagMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/distro.yaml", "name: my-distro\ndescription: A distro\ndevcontainer: arch-base@v1.0.0\npackages:\n  - shell-zsh@v1.3.0\n  - k8s-tools@v1.1.0\n")
	writeFile(t, dir+"/README.md", validDistroReadme)

	checker := &stubRefChecker{
		results: map[string]bool{packagesRepo + "@v1.3.0": false},
	}

	errs := validate.DistroDirWithOptions(dir, refOpts(checker))
	found := false
	for _, e := range errs {
		if strings.Contains(e, "shell-zsh@v1.3.0") && strings.Contains(e, "not found") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error for missing package tag, got: %v", errs)
	}
}

func TestRefCheck_multipleMissingPins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/distro.yaml", "name: my-distro\ndescription: A distro\ndevcontainer: arch-base@v2.0.0\npackages:\n  - shell-zsh@v1.3.0\n  - k8s-tools@v1.1.0\n")
	writeFile(t, dir+"/README.md", validDistroReadme)

	checker := &stubRefChecker{
		results: map[string]bool{
			devcontainerRepo + "@v2.0.0": false,
			packagesRepo + "@v1.3.0":     false,
		},
	}

	errs := validate.DistroDirWithOptions(dir, refOpts(checker))
	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors for 2 missing pins, got: %v", errs)
	}
}

func TestRefCheck_checkerError_propagated(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/distro.yaml", "name: my-distro\ndescription: A distro\ndevcontainer: arch-base@v2.0.0\npackages: []\n")
	writeFile(t, dir+"/README.md", validDistroReadme)

	checker := &stubRefChecker{
		errs: map[string]error{devcontainerRepo + "@v2.0.0": fmt.Errorf("network timeout")},
	}

	errs := validate.DistroDirWithOptions(dir, refOpts(checker))
	found := false
	for _, e := range errs {
		if strings.Contains(e, "network timeout") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error containing network timeout, got: %v", errs)
	}
}

func TestRefCheck_allExist_noErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/distro.yaml", "name: my-distro\ndescription: A distro\ndevcontainer: arch-base@v2.0.0\npackages:\n  - shell-zsh@v1.3.0\n")
	writeFile(t, dir+"/README.md", validDistroReadme)

	checker := &stubRefChecker{results: map[string]bool{}}

	errs := validate.DistroDirWithOptions(dir, refOpts(checker))
	if len(errs) != 0 {
		t.Errorf("expected no errors when all tags exist, got: %v", errs)
	}
}
