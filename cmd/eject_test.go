package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunEject_confirmedDeletesLock(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "dpod.lock")
	if err := os.WriteFile(lockPath, []byte("files: []"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runEject(lockPath, strings.NewReader("y\n"), &out); err != nil {
		t.Fatalf("runEject: %v", err)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("dpod.lock should have been deleted")
	}
	if !strings.Contains(out.String(), "Ejected") {
		t.Errorf("output should mention ejection, got: %s", out.String())
	}
}

func TestRunEject_abortPreservesLock(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "dpod.lock")
	if err := os.WriteFile(lockPath, []byte("files: []"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runEject(lockPath, strings.NewReader("n\n"), &out); err != nil {
		t.Fatalf("runEject: %v", err)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Error("dpod.lock should still exist after abort")
	}
	if !strings.Contains(out.String(), "Aborted") {
		t.Errorf("output should mention abort, got: %s", out.String())
	}
}

func TestRunEject_noLockReturnsError(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "dpod.lock")

	var out bytes.Buffer
	err := runEject(lockPath, strings.NewReader("y\n"), &out)
	if err == nil {
		t.Fatal("expected error when dpod.lock does not exist")
	}
	if !strings.Contains(err.Error(), "dpod.lock") {
		t.Errorf("error should mention dpod.lock, got: %v", err)
	}
}

func TestRunEject_emptyInputAborts(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "dpod.lock")
	if err := os.WriteFile(lockPath, []byte("files: []"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runEject(lockPath, strings.NewReader("\n"), &out); err != nil {
		t.Fatalf("runEject: %v", err)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Error("dpod.lock should still exist when Enter pressed with no input")
	}
}

func TestRunEject_promptText(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "dpod.lock")
	if err := os.WriteFile(lockPath, []byte("files: []"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	_ = runEject(lockPath, strings.NewReader("n\n"), &out)
	prompt := out.String()
	for _, want := range []string{"managed files", "untouched", "[y/N]"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q\n%s", want, prompt)
		}
	}
}
