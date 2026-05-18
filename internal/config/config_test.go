package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/config"
)

func TestRead_valid(t *testing.T) {
	yaml := `
distro: devops-k8s@v1.2.0
overrides:
  packages:
    add: [neovim]
    remove: [vscode]
  patches:
    - overrides/custom.patch
`
	path := writeTemp(t, yaml)
	cfg, err := config.Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Distro != "devops-k8s@v1.2.0" {
		t.Errorf("distro = %q, want %q", cfg.Distro, "devops-k8s@v1.2.0")
	}
	if len(cfg.Overrides.Packages.Add) != 1 || cfg.Overrides.Packages.Add[0] != "neovim" {
		t.Errorf("overrides.packages.add = %v, want [neovim]", cfg.Overrides.Packages.Add)
	}
	if len(cfg.Overrides.Packages.Remove) != 1 || cfg.Overrides.Packages.Remove[0] != "vscode" {
		t.Errorf("overrides.packages.remove = %v, want [vscode]", cfg.Overrides.Packages.Remove)
	}
	if len(cfg.Overrides.Patches) != 1 || cfg.Overrides.Patches[0] != "overrides/custom.patch" {
		t.Errorf("overrides.patches = %v, want [overrides/custom.patch]", cfg.Overrides.Patches)
	}
}

func TestRead_missingDistro(t *testing.T) {
	path := writeTemp(t, "overrides:\n  patches: []\n")
	_, err := config.Read(path)
	if err == nil {
		t.Fatal("expected error for missing distro field")
	}
}

func TestRead_malformedYAML(t *testing.T) {
	path := writeTemp(t, "distro: [invalid: yaml: {\n")
	_, err := config.Read(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestRead_missingFile(t *testing.T) {
	_, err := config.Read("/nonexistent/path/dpod.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWrite_roundTrip(t *testing.T) {
	original := &config.Config{
		Distro: "frontend-node@v0.3.1",
		Overrides: config.Overrides{
			Packages: config.PackageOverrides{
				Add: []string{"neovim"},
			},
			Patches: []string{"overrides/foo.patch"},
		},
	}

	path := filepath.Join(t.TempDir(), "dpod.yaml")
	if err := config.Write(path, original); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := config.Read(path)
	if err != nil {
		t.Fatalf("Read after Write: %v", err)
	}
	if got.Distro != original.Distro {
		t.Errorf("distro = %q, want %q", got.Distro, original.Distro)
	}
	if len(got.Overrides.Packages.Add) != 1 || got.Overrides.Packages.Add[0] != "neovim" {
		t.Errorf("packages.add = %v, want [neovim]", got.Overrides.Packages.Add)
	}
	if len(got.Overrides.Patches) != 1 || got.Overrides.Patches[0] != "overrides/foo.patch" {
		t.Errorf("patches = %v, want [overrides/foo.patch]", got.Overrides.Patches)
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dpod.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return path
}
