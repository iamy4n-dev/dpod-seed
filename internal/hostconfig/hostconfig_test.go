package hostconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/hostconfig"
)

func TestLoad_reposFromFile(t *testing.T) {
	content := `
repos:
  devcontainer: github.com/myorg/devcontainer
  packages: github.com/myorg/packages
  distro: github.com/myorg/distro
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := hostconfig.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Repos.Devcontainer != "github.com/myorg/devcontainer" {
		t.Errorf("Repos.Devcontainer = %q", cfg.Repos.Devcontainer)
	}
	if cfg.Repos.Packages != "github.com/myorg/packages" {
		t.Errorf("Repos.Packages = %q", cfg.Repos.Packages)
	}
	if cfg.Repos.Distro != "github.com/myorg/distro" {
		t.Errorf("Repos.Distro = %q", cfg.Repos.Distro)
	}
}

func TestLoad_registryURL(t *testing.T) {
	content := "registryUrl: https://raw.githubusercontent.com/myorg/distro/main/registry.yaml\n"
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := hostconfig.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := "https://raw.githubusercontent.com/myorg/distro/main/registry.yaml"
	if cfg.RegistryURL != want {
		t.Errorf("RegistryURL = %q, want %q", cfg.RegistryURL, want)
	}
}

func TestLoad_partialConfig(t *testing.T) {
	content := "repos:\n  distro: github.com/myorg/distro\n"
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := hostconfig.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Repos.Distro != "github.com/myorg/distro" {
		t.Errorf("Repos.Distro = %q", cfg.Repos.Distro)
	}
	if cfg.Repos.Devcontainer != "" {
		t.Errorf("Repos.Devcontainer should be empty, got %q", cfg.Repos.Devcontainer)
	}
	if cfg.RegistryURL != "" {
		t.Errorf("RegistryURL should be empty, got %q", cfg.RegistryURL)
	}
}

func TestLoad_missingFile(t *testing.T) {
	cfg, err := hostconfig.Load(filepath.Join(t.TempDir(), "config.yaml"))
	if err != nil {
		t.Fatalf("Load: expected no error for missing file, got %v", err)
	}
	if cfg.RegistryURL != "" || cfg.Repos.Distro != "" || cfg.Repos.Devcontainer != "" || cfg.Repos.Packages != "" {
		t.Errorf("expected zero-value Config, got %+v", cfg)
	}
}
