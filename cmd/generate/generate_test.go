package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/iamy4n-dev/dpod-seed/internal/generator"
)

const mockDistroYAML = `
packages:
  - shell-zsh@v1.3.0
  - k8s-tools@v1.1.0
`

func mockRegistryYAML(tag string) string {
	return fmt.Sprintf("distros:\n  - name: devops-k8s\n    description: Kubernetes development environment\n    latestTag: %s\n    status: stable\n", tag)
}

func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	var srv *httptest.Server
	srv = httptest.NewServer(mux)

	mux.HandleFunc("/registry.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, mockRegistryYAML("v0.2.0"))
	})
	mux.HandleFunc("/raw/distro.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, mockDistroYAML)
	})
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		downloadURL := srv.URL + "/raw/distro.yaml"
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"type\":\"file\",\"path\":\"distro.yaml\",\"download_url\":%q}", downloadURL)
	})
	return srv
}

func TestRunGenerate_writesPopulatedJSON(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Close()

	outPath := filepath.Join(t.TempDir(), "registry-data.json")
	cfg := generateConfig{
		registryURL: srv.URL + "/registry.yaml",
		distroRepo:  "github.com/iamy4n-dev/distros",
		githubBase:  srv.URL,
		outputPath:  outPath,
		httpClient:  srv.Client(),
	}

	if err := runGenerate(cfg); err != nil {
		t.Fatalf("runGenerate: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("output file not written: %v", err)
	}
	var out generator.Output
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(out.Distros) != 1 {
		t.Fatalf("expected 1 distro, got %d", len(out.Distros))
	}
	if out.Distros[0].Name != "devops-k8s" {
		t.Errorf("distro name = %q, want devops-k8s", out.Distros[0].Name)
	}
	if len(out.Distros[0].Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(out.Distros[0].Packages))
	}
}

func TestRunGenerate_fetchError_removesOutputFile(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/registry.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, mockRegistryYAML("v0.2.0"))
	})
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})

	outPath := filepath.Join(t.TempDir(), "registry-data.json")
	cfg := generateConfig{
		registryURL: srv.URL + "/registry.yaml",
		distroRepo:  "github.com/iamy4n-dev/distros",
		githubBase:  srv.URL,
		outputPath:  outPath,
		httpClient:  srv.Client(),
	}

	if err := runGenerate(cfg); err == nil {
		t.Fatal("expected error from fetch failure, got nil")
	}
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		t.Error("partial output file should be removed on error")
	}
}

func TestRunGenerate_withToken_sendsAuthHeader(t *testing.T) {
	const testToken = "ghp_testtoken123"
	var gotAuth string

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/registry.yaml", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		fmt.Fprint(w, mockRegistryYAML("v0.2.0"))
	})
	mux.HandleFunc("/raw/distro.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, mockDistroYAML)
	})
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		downloadURL := srv.URL + "/raw/distro.yaml"
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"type\":\"file\",\"path\":\"distro.yaml\",\"download_url\":%q}", downloadURL)
	})

	outPath := filepath.Join(t.TempDir(), "registry-data.json")
	cfg := generateConfig{
		registryURL: srv.URL + "/registry.yaml",
		distroRepo:  "github.com/iamy4n-dev/distros",
		githubBase:  srv.URL,
		outputPath:  outPath,
		httpClient:  newAuthClient(testToken, srv.Client().Transport),
	}

	if err := runGenerate(cfg); err != nil {
		t.Fatalf("runGenerate: %v", err)
	}
	want := "Bearer " + testToken
	if gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
}
