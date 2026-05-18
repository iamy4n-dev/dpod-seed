package fetch_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duyanh-y4n/dpod-seed/internal/fetch"
)

type ghEntry struct {
	Type        string `json:"type"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

func TestFetch_singleFile(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/org/repo/contents/profiles/arch-base":
			entries := []ghEntry{{
				Type:        "file",
				Path:        "profiles/arch-base/devcontainer.json",
				DownloadURL: srv.URL + "/raw/devcontainer.json",
			}}
			json.NewEncoder(w).Encode(entries)
		case "/raw/devcontainer.json":
			fmt.Fprint(w, `{"image":"mcr.microsoft.com/devcontainers/base:arch"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	f := fetch.NewGitHubFetcher(srv.URL, srv.Client())
	files, err := f.Fetch("github.com/org/repo", "abc123", "profiles/arch-base")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("files count = %d, want 1", len(files))
	}
	if files[0].Path != "devcontainer.json" {
		t.Errorf("path = %q, want %q", files[0].Path, "devcontainer.json")
	}
	if string(files[0].Content) != `{"image":"mcr.microsoft.com/devcontainers/base:arch"}` {
		t.Errorf("content = %q", string(files[0].Content))
	}
}

func TestFetch_directoryRecursive(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/org/repo/contents/dotfiles":
			entries := []ghEntry{
				{Type: "file", Path: "dotfiles/.zshrc", DownloadURL: srv.URL + "/raw/zshrc"},
				{Type: "dir", Path: "dotfiles/vim"},
			}
			json.NewEncoder(w).Encode(entries)
		case "/repos/org/repo/contents/dotfiles/vim":
			entries := []ghEntry{
				{Type: "file", Path: "dotfiles/vim/init.vim", DownloadURL: srv.URL + "/raw/init.vim"},
			}
			json.NewEncoder(w).Encode(entries)
		case "/raw/zshrc":
			fmt.Fprint(w, "export ZSH=$HOME/.oh-my-zsh")
		case "/raw/init.vim":
			fmt.Fprint(w, `set number`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	f := fetch.NewGitHubFetcher(srv.URL, srv.Client())
	files, err := f.Fetch("github.com/org/repo", "sha1", "dotfiles")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("files count = %d, want 2", len(files))
	}
}

func TestFetch_non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	f := fetch.NewGitHubFetcher(srv.URL, srv.Client())
	_, err := f.Fetch("github.com/org/repo", "badshaXXX", "some/path")
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestFetch_networkFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately

	f := fetch.NewGitHubFetcher(srv.URL, srv.Client())
	_, err := f.Fetch("github.com/org/repo", "sha1", "path")
	if err == nil {
		t.Fatal("expected error for network failure")
	}
}
