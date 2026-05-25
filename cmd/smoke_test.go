package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iamy4n-dev/dpod-seed/internal/generator"
	"github.com/iamy4n-dev/dpod-seed/internal/resolver"
)

// TestSmoke_InitSyncValidatePipeline exercises the full user-facing loop:
// init selects a distro and runs sync; a second sync is idempotent; validate passes.
func TestSmoke_InitSyncValidatePipeline(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	distros := []generator.DistroRecord{
		{Name: "example", Description: "placeholder", LatestTag: "v0.1.0",
			Devcontainer: "arch-base@v0.1.0",
			Packages:     []generator.Package{{Name: "cli-essentials", Version: "v0.1.0"}}},
	}
	res := &mockResolver{entries: []resolver.ManifestEntry{
		{DestPath: ".devcontainer/devcontainer.json", SrcRepo: "github.com/x/dc", SHA: "abc123", Content: []byte(`{}`)},
	}}

	// --- init: select distro 1, confirm ---
	var out bytes.Buffer
	err := runInit(strings.NewReader("1\ny\n"), &out, true, configPath, lockPath, dir, distros, res)
	if err != nil {
		t.Fatalf("runInit: %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("dpod.yaml not written after init: %v", err)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("dpod.lock not written after init: %v", err)
	}

	// --- second sync: must be idempotent ---
	out.Reset()
	if err := runSync(&out, configPath, lockPath, dir, res); err != nil {
		t.Fatalf("runSync (second): %v", err)
	}
	if !strings.Contains(out.String(), "0 added, 0 updated, 0 removed") {
		t.Errorf("second sync should be idempotent, got: %q", out.String())
	}

	// --- validate: must pass on materialised state ---
	out.Reset()
	if err := runValidate(configPath, &out, res); err != nil {
		t.Fatalf("runValidate: %v", err)
	}
	if !strings.Contains(out.String(), "OK") {
		t.Errorf("validate should report OK, got: %q", out.String())
	}
}

func TestSmoke_VersionIsDevByDefault(t *testing.T) {
	if version != "dev" {
		t.Errorf("version should be %q in dev builds, got %q — release ldflag is being applied unexpectedly", "dev", version)
	}
}
