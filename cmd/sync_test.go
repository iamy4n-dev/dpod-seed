package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/resolver"
)

func TestRunSync_missingConfigReturnsError(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := runSync(&out, filepath.Join(dir, "dpod.yaml"), filepath.Join(dir, "dpod.lock"), dir, &mockResolver{})
	if err == nil {
		t.Fatal("expected error for missing dpod.yaml")
	}
}

func TestRunSync_resolverErrorLeavesLockUntouched(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	if err := os.WriteFile(configPath, []byte("distro: devops-k8s@v1.2.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := &mockResolver{err: errors.New("distro not found")}
	var out bytes.Buffer
	if err := runSync(&out, configPath, lockPath, dir, res); err == nil {
		t.Fatal("expected error from resolver")
	}

	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("dpod.lock should not be written when resolver fails")
	}
}

func TestRunSync_incrementalNoOpSkipsWrite(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	if err := os.WriteFile(configPath, []byte("distro: devops-k8s@v1.2.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// pre-seed the file and the lock with matching SHA
	dest := filepath.Join(dir, ".devcontainer/devcontainer.json")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte(`{"name":"original"}`)
	if err := os.WriteFile(dest, original, 0o644); err != nil {
		t.Fatal(err)
	}
	lockContent := "files:\n  - path: .devcontainer/devcontainer.json\n    repo: github.com/x/dc\n    sha: abc123\n"
	if err := os.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatal(err)
	}

	res := &mockResolver{entries: []resolver.ManifestEntry{
		{DestPath: ".devcontainer/devcontainer.json", SrcRepo: "github.com/x/dc", SHA: "abc123", Content: []byte(`{"name":"new"}`)},
	}}

	var out bytes.Buffer
	if err := runSync(&out, configPath, lockPath, dir, res); err != nil {
		t.Fatalf("runSync: %v", err)
	}

	// file should NOT have been overwritten
	data, _ := os.ReadFile(dest)
	if string(data) != string(original) {
		t.Errorf("file should not be re-written on no-op, got: %s", data)
	}

	if !strings.Contains(out.String(), "0 added, 0 updated, 0 removed") {
		t.Errorf("unexpected summary: %q", out.String())
	}
}

func TestRunSync_updatedFileIsOverwritten(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	if err := os.WriteFile(configPath, []byte("distro: devops-k8s@v1.2.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, ".devcontainer/devcontainer.json")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte(`{"name":"old"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// lock has old SHA
	lockContent := "files:\n  - path: .devcontainer/devcontainer.json\n    repo: github.com/x/dc\n    sha: oldsha\n"
	if err := os.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatal(err)
	}

	res := &mockResolver{entries: []resolver.ManifestEntry{
		{DestPath: ".devcontainer/devcontainer.json", SrcRepo: "github.com/x/dc", SHA: "newsha", Content: []byte(`{"name":"new"}`)},
	}}

	var out bytes.Buffer
	if err := runSync(&out, configPath, lockPath, dir, res); err != nil {
		t.Fatalf("runSync: %v", err)
	}

	data, _ := os.ReadFile(dest)
	if string(data) != `{"name":"new"}` {
		t.Errorf("file should be overwritten, got: %s", data)
	}
	if !strings.Contains(out.String(), "0 added, 1 updated, 0 removed") {
		t.Errorf("unexpected summary: %q", out.String())
	}
}

func TestRunSync_removedFileIsDeletedFromDisk(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	if err := os.WriteFile(configPath, []byte("distro: devops-k8s@v1.2.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// file that will be dropped from the new manifest
	stale := filepath.Join(dir, ".devcontainer/old.json")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	lockContent := "files:\n  - path: .devcontainer/old.json\n    repo: github.com/x/dc\n    sha: abc123\n"
	if err := os.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// new manifest no longer contains old.json
	res := &mockResolver{entries: []resolver.ManifestEntry{
		{DestPath: ".devcontainer/devcontainer.json", SrcRepo: "github.com/x/dc", SHA: "newsha", Content: []byte(`{"name":"new"}`)},
	}}

	var out bytes.Buffer
	if err := runSync(&out, configPath, lockPath, dir, res); err != nil {
		t.Fatalf("runSync: %v", err)
	}

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("stale file should have been deleted from disk")
	}
	if !strings.Contains(out.String(), "1 added, 0 updated, 1 removed") {
		t.Errorf("unexpected summary: %q", out.String())
	}
}

func TestRunSync_materializerFailureLeavesLockUntouched(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	if err := os.WriteFile(configPath, []byte("distro: devops-k8s@v1.2.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// dest is a directory, not a file — WriteFile will fail
	conflictDir := filepath.Join(dir, ".devcontainer", "devcontainer.json")
	if err := os.MkdirAll(conflictDir, 0o755); err != nil {
		t.Fatal(err)
	}

	res := &mockResolver{entries: []resolver.ManifestEntry{
		{DestPath: ".devcontainer/devcontainer.json", SrcRepo: "github.com/x/dc", SHA: "abc123", Content: []byte(`{}`)},
	}}

	var out bytes.Buffer
	if err := runSync(&out, configPath, lockPath, dir, res); err == nil {
		t.Fatal("expected error when materializer fails")
	}

	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("dpod.lock should not be written when materializer fails")
	}
}

func TestRunSync_freshSyncWritesFileAndLock(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	if err := os.WriteFile(configPath, []byte("distro: devops-k8s@v1.2.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := &mockResolver{entries: []resolver.ManifestEntry{
		{DestPath: ".devcontainer/devcontainer.json", SrcRepo: "github.com/x/dc", SHA: "abc123", Content: []byte(`{"name":"test"}`)},
	}}

	var out bytes.Buffer
	if err := runSync(&out, configPath, lockPath, dir, res); err != nil {
		t.Fatalf("runSync: %v", err)
	}

	// file materialised on disk
	written := filepath.Join(dir, ".devcontainer/devcontainer.json")
	data, err := os.ReadFile(written)
	if err != nil {
		t.Fatalf("expected file on disk: %v", err)
	}
	if string(data) != `{"name":"test"}` {
		t.Errorf("unexpected file content: %s", data)
	}

	// lock written
	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("dpod.lock not written: %v", err)
	}

	// summary printed
	if !strings.Contains(out.String(), "1 added, 0 updated, 0 removed") {
		t.Errorf("unexpected summary: %q", out.String())
	}
}
