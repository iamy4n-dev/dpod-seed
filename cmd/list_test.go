package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/registry"
)

type mockRegistryClient struct {
	entries []registry.DistroEntry
	err     error
}

func (m *mockRegistryClient) List() ([]registry.DistroEntry, error) {
	return m.entries, m.err
}

func TestRunList_printsTable(t *testing.T) {
	client := &mockRegistryClient{
		entries: []registry.DistroEntry{
			{Name: "devops-k8s", Description: "Arch with K8s tooling", LatestTag: "v1.2.0"},
			{Name: "frontend-node", Description: "Node.js environment", LatestTag: "v0.3.1"},
		},
	}
	var buf bytes.Buffer
	if err := runList(&buf, client); err != nil {
		t.Fatalf("runList: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"devops-k8s", "v1.2.0", "Arch with K8s tooling", "frontend-node", "v0.3.1"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestRunList_registryError(t *testing.T) {
	client := &mockRegistryClient{err: errors.New("connection refused")}
	var buf bytes.Buffer
	if err := runList(&buf, client); err == nil {
		t.Fatal("expected error from registry client")
	}
}
