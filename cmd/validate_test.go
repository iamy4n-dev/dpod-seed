package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/config"
	"github.com/duyanh-y4n/dpod-seed/internal/resolver"
)

type mockResolver struct {
	entries []resolver.ManifestEntry
	err     error
}

func (m *mockResolver) Resolve(distroName, tag string, overrides config.Overrides) ([]resolver.ManifestEntry, error) {
	return m.entries, m.err
}

func TestRunValidate_validConfig(t *testing.T) {
	path := tempConfig(t, "distro: devops-k8s@v1.2.0\n")
	r := &mockResolver{entries: []resolver.ManifestEntry{{DestPath: ".devcontainer/devcontainer.json"}}}

	var out bytes.Buffer
	if err := runValidate(path, &out, r); err != nil {
		t.Fatalf("runValidate: %v", err)
	}
	if !strings.Contains(out.String(), "OK") {
		t.Errorf("output should contain OK, got: %s", out.String())
	}
}

func TestRunValidate_missingConfig(t *testing.T) {
	var out bytes.Buffer
	err := runValidate("/nonexistent/dpod.yaml", &out, &mockResolver{})
	if err == nil {
		t.Fatal("expected error for missing dpod.yaml")
	}
}

func TestRunValidate_resolveFailure(t *testing.T) {
	path := tempConfig(t, "distro: bad-distro@v0.0.0\n")
	r := &mockResolver{err: errors.New("distro not found")}

	var out bytes.Buffer
	err := runValidate(path, &out, r)
	if err == nil {
		t.Fatal("expected error when resolve fails")
	}
	if !strings.Contains(err.Error(), "distro not found") {
		t.Errorf("error should propagate resolver message, got: %v", err)
	}
}

func tempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dpod.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("tempConfig: %v", err)
	}
	return path
}
