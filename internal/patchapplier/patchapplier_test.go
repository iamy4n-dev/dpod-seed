package patchapplier_test

import (
	"strings"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/patchapplier"
)

func TestApply_replaceLine(t *testing.T) {
	src := "hello\nworld\n"
	patch := "--- a/file\n+++ b/file\n@@ -2,1 +2,1 @@\n-world\n+earth\n"
	got, err := patchapplier.Apply([]byte(src), []byte(patch))
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if string(got) != "hello\nearth\n" {
		t.Errorf("got %q, want %q", got, "hello\nearth\n")
	}
}

func TestApply_addLine(t *testing.T) {
	src := "line1\nline2\n"
	patch := "--- a/file\n+++ b/file\n@@ -1,2 +1,3 @@\n line1\n+new line\n line2\n"
	got, err := patchapplier.Apply([]byte(src), []byte(patch))
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if string(got) != "line1\nnew line\nline2\n" {
		t.Errorf("got %q, want %q", got, "line1\nnew line\nline2\n")
	}
}

func TestApply_removeLine(t *testing.T) {
	src := "a\nb\nc\n"
	patch := "--- a/file\n+++ b/file\n@@ -1,3 +1,2 @@\n a\n-b\n c\n"
	got, err := patchapplier.Apply([]byte(src), []byte(patch))
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if string(got) != "a\nc\n" {
		t.Errorf("got %q, want %q", got, "a\nc\n")
	}
}

func TestApply_malformedHunkHeader(t *testing.T) {
	patch := "--- a/file\n+++ b/file\n@@ not a hunk @@\n line\n"
	_, err := patchapplier.Apply([]byte("line\n"), []byte(patch))
	if err == nil {
		t.Fatal("expected error for malformed hunk header")
	}
}

func TestApply_contextMismatch(t *testing.T) {
	src := "hello\nworld\n"
	// patch expects "universe" as context but source has "world"
	patch := "--- a/file\n+++ b/file\n@@ -1,2 +1,2 @@\n hello\n-universe\n+earth\n"
	_, err := patchapplier.Apply([]byte(src), []byte(patch))
	if err == nil {
		t.Fatal("expected error for context mismatch")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("error should mention mismatch, got: %v", err)
	}
}
