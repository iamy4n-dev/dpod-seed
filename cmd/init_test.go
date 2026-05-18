package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/registry"
	"github.com/duyanh-y4n/dpod-seed/internal/resolver"
)

// twoDistros is a reusable registry client used across init tests.
var twoDistros = &mockRegistryClient{
	entries: []registry.DistroEntry{
		{Name: "devops-k8s", Description: "K8s tooling", LatestTag: "v1.0.0"},
		{Name: "frontend-node", Description: "Node.js env", LatestTag: "v0.5.0"},
	},
}

// oneFile is a reusable resolver used across init tests.
var oneFile = &mockResolver{entries: []resolver.ManifestEntry{
	{DestPath: ".devcontainer/devcontainer.json", SrcRepo: "github.com/x/dc", SHA: "abc", Content: []byte(`{}`)},
}}

func TestRunInit_noTTYReturnsError(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := runInit(strings.NewReader(""), &out, false,
		filepath.Join(dir, "dpod.yaml"), filepath.Join(dir, "dpod.lock"), dir,
		twoDistros, oneFile)
	if err == nil {
		t.Fatal("expected error when no TTY")
	}
	msg := err.Error()
	if !strings.Contains(msg, "dpod.yaml") || !strings.Contains(msg, "dpod-seed sync") {
		t.Errorf("error should mention dpod.yaml and dpod-seed sync, got: %s", msg)
	}
}

func TestRunInit_cancelWritesNothing(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	input := strings.NewReader("1\nc\n")
	var out bytes.Buffer
	if err := runInit(input, &out, true, configPath, lockPath, dir, twoDistros, oneFile); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("dpod.yaml should not be written on cancel")
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("dpod.lock should not be written on cancel")
	}
}

func TestRunInit_retryThenConfirmUsesSecondPick(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	// pick 1, retry, pick 2, confirm
	input := strings.NewReader("1\nr\n2\ny\n")
	var out bytes.Buffer
	if err := runInit(input, &out, true, configPath, lockPath, dir, twoDistros, oneFile); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("dpod.yaml not written: %v", err)
	}
	if !strings.Contains(string(data), "frontend-node") {
		t.Errorf("dpod.yaml should contain second distro, got: %s", data)
	}
	if strings.Contains(string(data), "devops-k8s") {
		t.Errorf("dpod.yaml should not contain first distro after retry, got: %s", data)
	}
}

func TestRunInit_existingConfigOverwriteRefused(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	original := []byte("distro: old-distro@v0.0.1\n")
	if err := os.WriteFile(configPath, original, 0o644); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("n\n")
	var out bytes.Buffer
	if err := runInit(input, &out, true, configPath, lockPath, dir, twoDistros, oneFile); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	if string(data) != string(original) {
		t.Errorf("dpod.yaml should be unchanged after overwrite refusal, got: %s", data)
	}
	if !strings.Contains(out.String(), "already exists") {
		t.Errorf("output should warn about existing file, got: %s", out.String())
	}
}

func TestRunInit_existingConfigOverwriteConfirmed(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	if err := os.WriteFile(configPath, []byte("distro: old-distro@v0.0.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// confirm overwrite, select distro 1, confirm sync
	input := strings.NewReader("y\n1\ny\n")
	var out bytes.Buffer
	if err := runInit(input, &out, true, configPath, lockPath, dir, twoDistros, oneFile); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("dpod.yaml not found: %v", err)
	}
	if !strings.Contains(string(data), "devops-k8s") {
		t.Errorf("dpod.yaml should be overwritten with new distro, got: %s", data)
	}
}

func TestRunInit_registryErrorPropagated(t *testing.T) {
	dir := t.TempDir()
	client := &mockRegistryClient{err: errors.New("connection refused")}
	var out bytes.Buffer
	err := runInit(strings.NewReader(""), &out, true,
		filepath.Join(dir, "dpod.yaml"), filepath.Join(dir, "dpod.lock"), dir,
		client, oneFile)
	if err == nil {
		t.Fatal("expected error from registry client")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error should propagate registry message, got: %v", err)
	}
}

func TestRunInit_happyPath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "dpod.yaml")
	lockPath := filepath.Join(dir, "dpod.lock")

	// input: select distro 1, confirm
	input := strings.NewReader("1\ny\n")
	var out bytes.Buffer

	if err := runInit(input, &out, true, configPath, lockPath, dir, twoDistros, oneFile); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// distro list shown
	outStr := out.String()
	if !strings.Contains(outStr, "devops-k8s") {
		t.Errorf("distro list not shown, output: %s", outStr)
	}

	// dpod.yaml written with selected distro
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("dpod.yaml not written: %v", err)
	}
	if !strings.Contains(string(data), "devops-k8s") {
		t.Errorf("dpod.yaml should contain selected distro, got: %s", data)
	}

	// lock written (sync ran)
	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("dpod.lock not written: %v", err)
	}
}
