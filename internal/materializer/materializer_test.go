package materializer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/lock"
	"github.com/duyanh-y4n/dpod-seed/internal/materializer"
	"github.com/duyanh-y4n/dpod-seed/internal/resolver"
)

func TestMaterialize_writesFileToDestPath(t *testing.T) {
	dir := t.TempDir()
	entries := []resolver.ManifestEntry{
		{DestPath: ".devcontainer/.zshrc", SrcRepo: "github.com/org/packages", SHA: "v1.0.0", Content: []byte("export ZSH=$HOME/.oh-my-zsh")},
	}
	summary, err := materializer.Materialize(dir, entries, nil, nil)
	if err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".devcontainer/.zshrc"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(got) != "export ZSH=$HOME/.oh-my-zsh" {
		t.Errorf("content = %q", got)
	}
	if summary.Written != 1 {
		t.Errorf("Written = %d, want 1", summary.Written)
	}
}

func TestMaterialize_createsParentDirs(t *testing.T) {
	dir := t.TempDir()
	entries := []resolver.ManifestEntry{
		{DestPath: ".devcontainer/nested/deep/config", Content: []byte("data")},
	}
	if _, err := materializer.Materialize(dir, entries, nil, nil); err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".devcontainer/nested/deep/config")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestMaterialize_deletesRemovedFiles(t *testing.T) {
	dir := t.TempDir()
	// pre-create a file that should be removed
	removePath := filepath.Join(dir, ".devcontainer/old.conf")
	if err := os.MkdirAll(filepath.Dir(removePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(removePath, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	removes := []lock.File{{Path: ".devcontainer/old.conf", Repo: "github.com/org/packages", SHA: "v1.0.0"}}
	summary, err := materializer.Materialize(dir, nil, removes, nil)
	if err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	if _, err := os.Stat(removePath); !os.IsNotExist(err) {
		t.Error(".devcontainer/old.conf should have been deleted")
	}
	if summary.Removed != 1 {
		t.Errorf("Removed = %d, want 1", summary.Removed)
	}
}

func TestMaterialize_appliesPatchFile(t *testing.T) {
	dir := t.TempDir()
	// pre-create target file
	targetPath := filepath.Join(dir, ".devcontainer/devcontainer.json")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetPath, []byte(`{"image":"base","extensions":[]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// write patch file
	patchContent := "--- a/.devcontainer/devcontainer.json\n+++ b/.devcontainer/devcontainer.json\n@@ -1,1 +1,1 @@\n-{\"image\":\"base\",\"extensions\":[]}\n+{\"image\":\"base\",\"extensions\":[\"ms-python.python\"]}\n"
	patchDir := filepath.Join(dir, "overrides")
	if err := os.MkdirAll(patchDir, 0o755); err != nil {
		t.Fatal(err)
	}
	patchFile := filepath.Join(patchDir, "custom.patch")
	if err := os.WriteFile(patchFile, []byte(patchContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := materializer.Materialize(dir, nil, nil, []string{"overrides/custom.patch"}); err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	got, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(got) != `{"image":"base","extensions":["ms-python.python"]}`+"\n" {
		t.Errorf("patched content = %q", got)
	}
}

func TestMaterialize_missingRemoveFileIsIgnored(t *testing.T) {
	dir := t.TempDir()
	removes := []lock.File{{Path: ".devcontainer/gone.conf", Repo: "r", SHA: "s"}}
	// should not error if file to remove is already absent
	if _, err := materializer.Materialize(dir, nil, removes, nil); err != nil {
		t.Fatalf("Materialize: %v", err)
	}
}
