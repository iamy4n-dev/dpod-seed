package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldDistro_createsFiles(t *testing.T) {
	base := t.TempDir()
	if err := runScaffoldDistro("my-distro", base); err != nil {
		t.Fatalf("runScaffoldDistro: %v", err)
	}

	distroYAML := filepath.Join(base, "distros", "my-distro", "distro.yaml")
	if _, err := os.Stat(distroYAML); err != nil {
		t.Errorf("distro.yaml not created: %v", err)
	}
	content, _ := os.ReadFile(distroYAML)
	if !strings.Contains(string(content), "my-distro") {
		t.Error("distro.yaml should contain the distro name")
	}

	readme := filepath.Join(base, "distros", "my-distro", "README.md")
	content, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("README.md not created: %v", err)
	}
	for _, want := range []string{
		"---",
		"display_name:", "description:", "status: experimental",
		"devcontainer:", "tags:", "packages:",
		"```yaml", "dpod.yaml",
	} {
		if !strings.Contains(string(content), want) {
			t.Errorf("distro README missing %q", want)
		}
	}
}

func TestScaffoldDistro_existingPathErrors(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "distros", "my-distro")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := runScaffoldDistro("my-distro", base); err == nil {
		t.Fatal("expected error for existing path")
	}
}

func TestScaffoldPackage_createsFiles(t *testing.T) {
	base := t.TempDir()
	if err := runScaffoldPackage("my-pkg", base); err != nil {
		t.Fatalf("runScaffoldPackage: %v", err)
	}

	dotfiles := filepath.Join(base, "packages", "my-pkg", "dotfiles")
	if info, err := os.Stat(dotfiles); err != nil || !info.IsDir() {
		t.Errorf("dotfiles/ directory not created")
	}

	manifest := filepath.Join(base, "packages", "my-pkg", "manifest.yaml")
	if _, err := os.Stat(manifest); err != nil {
		t.Errorf("manifest.yaml not created: %v", err)
	}
	content, _ := os.ReadFile(manifest)
	if !strings.Contains(string(content), "dotfiles/") {
		t.Error("manifest.yaml should document dotfiles/ convention")
	}

	pkgReadme := filepath.Join(base, "packages", "my-pkg", "README.md")
	pkgContent, err := os.ReadFile(pkgReadme)
	if err != nil {
		t.Fatalf("README.md not created: %v", err)
	}
	for _, want := range []string{
		"---",
		"display_name:", "description:", "status: experimental",
		"tags:", "tools:",
	} {
		if !strings.Contains(string(pkgContent), want) {
			t.Errorf("package README missing %q", want)
		}
	}
}

func TestScaffoldPackage_existingPathErrors(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "packages", "my-pkg")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := runScaffoldPackage("my-pkg", base); err == nil {
		t.Fatal("expected error for existing path")
	}
}
