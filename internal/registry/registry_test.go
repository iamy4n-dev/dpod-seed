package registry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/registry"
)

const validRegistry = `
distros:
  - name: devops-k8s
    description: Arch-based OS with AWS and Kubernetes tooling
    latestTag: v1.2.0
    changelogUrl: https://example.com/v1.2.0
  - name: frontend-node
    description: Node.js frontend environment
    latestTag: v0.3.1
    changelogUrl: https://example.com/v0.3.1
`

func TestList_valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(validRegistry))
	}))
	defer srv.Close()

	c := registry.NewClient(srv.URL, srv.Client())
	entries, err := c.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries count = %d, want 2", len(entries))
	}
	if entries[0].Name != "devops-k8s" {
		t.Errorf("entries[0].name = %q, want %q", entries[0].Name, "devops-k8s")
	}
	if entries[0].LatestTag != "v1.2.0" {
		t.Errorf("entries[0].latestTag = %q, want %q", entries[0].LatestTag, "v1.2.0")
	}
	if entries[1].Name != "frontend-node" {
		t.Errorf("entries[1].name = %q", entries[1].Name)
	}
}

func TestList_non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := registry.NewClient(srv.URL, srv.Client())
	_, err := c.List()
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestList_invalidYAML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("distros: [invalid: yaml: {\n"))
	}))
	defer srv.Close()

	c := registry.NewClient(srv.URL, srv.Client())
	_, err := c.List()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestList_networkFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := registry.NewClient(srv.URL, srv.Client())
	_, err := c.List()
	if err == nil {
		t.Fatal("expected error for network failure")
	}
}
