package lock_test

import (
	"path/filepath"
	"testing"

	"github.com/iamy4n-dev/dpod-seed/internal/lock"
)

func TestRead_nonExistentFile(t *testing.T) {
	l, err := lock.Read("/nonexistent/dpod.lock")
	if err != nil {
		t.Fatalf("expected empty lock, got error: %v", err)
	}
	if len(l.Files) != 0 {
		t.Errorf("expected empty files, got %d", len(l.Files))
	}
}

func TestWrite_roundTrip(t *testing.T) {
	original := &lock.Lock{
		Files: []lock.File{
			{Path: ".devcontainer/devcontainer.json", Repo: "github.com/org/devcontainer", SHA: "abc123"},
			{Path: ".devcontainer/dotfiles/.zshrc", Repo: "github.com/org/packages", SHA: "def456"},
		},
	}

	path := filepath.Join(t.TempDir(), "dpod.lock")
	if err := lock.Write(path, original); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := lock.Read(path)
	if err != nil {
		t.Fatalf("Read after Write: %v", err)
	}
	if len(got.Files) != 2 {
		t.Fatalf("files count = %d, want 2", len(got.Files))
	}
	if got.Files[0].Path != ".devcontainer/devcontainer.json" {
		t.Errorf("files[0].path = %q", got.Files[0].Path)
	}
	if got.Files[1].SHA != "def456" {
		t.Errorf("files[1].sha = %q", got.Files[1].SHA)
	}
}

func TestCompute_emptyOld(t *testing.T) {
	newFiles := []lock.File{
		{Path: "a", Repo: "r", SHA: "1"},
		{Path: "b", Repo: "r", SHA: "2"},
	}
	d := lock.Compute(&lock.Lock{}, newFiles)

	if len(d.Added) != 2 {
		t.Errorf("added = %d, want 2", len(d.Added))
	}
	if len(d.Updated) != 0 {
		t.Errorf("updated = %d, want 0", len(d.Updated))
	}
	if len(d.Removed) != 0 {
		t.Errorf("removed = %d, want 0", len(d.Removed))
	}
}

func TestCompute_shaChanged(t *testing.T) {
	old := &lock.Lock{Files: []lock.File{{Path: "a", Repo: "r", SHA: "old"}}}
	newFiles := []lock.File{{Path: "a", Repo: "r", SHA: "new"}}
	d := lock.Compute(old, newFiles)

	if len(d.Updated) != 1 || d.Updated[0].SHA != "new" {
		t.Errorf("updated = %v, want [{a r new}]", d.Updated)
	}
	if len(d.Added) != 0 || len(d.Removed) != 0 {
		t.Errorf("expected no added/removed, got added=%v removed=%v", d.Added, d.Removed)
	}
}

func TestCompute_removedFiles(t *testing.T) {
	old := &lock.Lock{Files: []lock.File{
		{Path: "a", Repo: "r", SHA: "1"},
		{Path: "b", Repo: "r", SHA: "2"},
	}}
	newFiles := []lock.File{{Path: "a", Repo: "r", SHA: "1"}}
	d := lock.Compute(old, newFiles)

	if len(d.Removed) != 1 || d.Removed[0].Path != "b" {
		t.Errorf("removed = %v, want [{b r 2}]", d.Removed)
	}
	if len(d.Added) != 0 || len(d.Updated) != 0 {
		t.Errorf("expected no added/updated, got added=%v updated=%v", d.Added, d.Updated)
	}
}

func TestCompute_unchangedFiles(t *testing.T) {
	files := []lock.File{{Path: "a", Repo: "r", SHA: "1"}}
	old := &lock.Lock{Files: files}
	d := lock.Compute(old, files)

	if len(d.Added) != 0 || len(d.Updated) != 0 || len(d.Removed) != 0 {
		t.Errorf("expected empty diff, got added=%v updated=%v removed=%v", d.Added, d.Updated, d.Removed)
	}
}
