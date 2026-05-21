package validate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iamy4n-dev/dpod-seed/internal/validate"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const validDistroYAML = `name: devops-k8s
description: Kubernetes development environment
devcontainer: arch-base@v2.0.0
packages:
  - shell-zsh@v1.3.0
  - k8s-tools@v1.1.0
`

const validDistroReadme = `---
name: "devops-k8s"
display_name: "DevOps K8s"
description: "Kubernetes development environment"
status: stable
devcontainer: arch-base@v2.0.0
tags: [k8s, devops]
packages:
  - shell-zsh@v1.3.0
---

# devops-k8s

Some description.
`

func TestValidateDistroDir_missingFields(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "distros", "bad")
	writeFile(t, filepath.Join(dir, "distro.yaml"), `name: ""
description: ""
devcontainer: ""
packages: []
`)
	writeFile(t, filepath.Join(dir, "README.md"), validDistroReadme)

	errs := validate.DistroDir(dir)
	want := []string{"name", "description", "devcontainer"}
	for _, w := range want {
		found := false
		for _, e := range errs {
			if strings.Contains(e, w) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected error mentioning %q, got: %v", w, errs)
		}
	}
}

func TestValidateDistroDir_badPin(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "distros", "bad-pin")
	writeFile(t, filepath.Join(dir, "distro.yaml"), `name: my-distro
description: A distro
devcontainer: arch-base-no-version
packages:
  - shell-zsh@notaversion
`)
	writeFile(t, filepath.Join(dir, "README.md"), validDistroReadme)

	errs := validate.DistroDir(dir)
	if len(errs) != 2 {
		t.Errorf("expected 2 pin errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateDistroDir_badReadmeStatus(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "distros", "bad-status")
	writeFile(t, filepath.Join(dir, "distro.yaml"), validDistroYAML)
	writeFile(t, filepath.Join(dir, "README.md"), `---
name: "my-distro"
description: "A distro"
status: draft
---

# my-distro
`)

	errs := validate.DistroDir(dir)
	if len(errs) != 1 || !strings.Contains(errs[0], "status") {
		t.Errorf("expected one status error, got: %v", errs)
	}
}

func TestValidateDistroDir_missingReadmeFrontmatter(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "distros", "no-fm")
	writeFile(t, filepath.Join(dir, "distro.yaml"), validDistroYAML)
	writeFile(t, filepath.Join(dir, "README.md"), "# my-distro\n\nNo frontmatter here.\n")

	errs := validate.DistroDir(dir)
	if len(errs) == 0 {
		t.Error("expected error for missing frontmatter")
	}
}

func TestValidatePackageDir_valid(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "packages", "my-pkg")
	writeFile(t, filepath.Join(dir, "manifest.yaml"), "files: []\n")
	writeFile(t, filepath.Join(dir, "README.md"), `---
name: "my-pkg"
description: "A package bundle"
status: experimental
tags: []
tools: []
---

# my-pkg
`)

	errs := validate.PackageDir(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidatePackageDir_missingFields(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "packages", "bad-pkg")
	writeFile(t, filepath.Join(dir, "manifest.yaml"), "files: []\n")
	writeFile(t, filepath.Join(dir, "README.md"), `---
name: ""
description: ""
status: stable
---

# bad-pkg
`)

	errs := validate.PackageDir(dir)
	want := []string{"name", "description"}
	for _, w := range want {
		found := false
		for _, e := range errs {
			if strings.Contains(e, w) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected error mentioning %q, got: %v", w, errs)
		}
	}
}

func TestRun_distrosRepo(t *testing.T) {
	base := t.TempDir()
	dir1 := filepath.Join(base, "distros", "good-distro")
	writeFile(t, filepath.Join(dir1, "distro.yaml"), validDistroYAML)
	writeFile(t, filepath.Join(dir1, "README.md"), validDistroReadme)

	dir2 := filepath.Join(base, "distros", "bad-distro")
	writeFile(t, filepath.Join(dir2, "distro.yaml"), `name: ""
description: ""
devcontainer: arch-base@v1.0.0
packages: []
`)
	writeFile(t, filepath.Join(dir2, "README.md"), validDistroReadme)

	errs := validate.Run(base)
	if len(errs) == 0 {
		t.Error("expected errors from bad-distro, got none")
	}
	for _, e := range errs {
		if strings.Contains(e, "good-distro") {
			t.Errorf("unexpected error for good-distro: %s", e)
		}
	}
}

func TestRun_packagesRepo(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "packages", "my-pkg")
	writeFile(t, filepath.Join(dir, "manifest.yaml"), "files: []\n")
	writeFile(t, filepath.Join(dir, "README.md"), `---
name: "my-pkg"
description: "A package bundle"
status: experimental
---

# my-pkg
`)

	errs := validate.Run(base)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestRun_noRecognisedLayout(t *testing.T) {
	base := t.TempDir()
	errs := validate.Run(base)
	if len(errs) == 0 {
		t.Error("expected error when no distros/ or packages/ directory found")
	}
}

func TestValidateDistroDir_valid(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "distros", "devops-k8s")
	writeFile(t, filepath.Join(dir, "distro.yaml"), validDistroYAML)
	writeFile(t, filepath.Join(dir, "README.md"), validDistroReadme)

	errs := validate.DistroDir(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}
