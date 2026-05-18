package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type File struct {
	Path    string
	Content []byte
}

type Fetcher interface {
	Fetch(repo, sha, path string) ([]File, error)
}

type githubFetcher struct {
	baseURL    string
	httpClient *http.Client
}

func NewGitHubFetcher(baseURL string, httpClient *http.Client) Fetcher {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &githubFetcher{baseURL: baseURL, httpClient: httpClient}
}

type ghEntry struct {
	Type        string `json:"type"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

func (f *githubFetcher) Fetch(repo, sha, basePath string) ([]File, error) {
	ownerRepo := strings.TrimPrefix(repo, "github.com/")
	return f.walk(ownerRepo, sha, basePath, basePath)
}

func (f *githubFetcher) walk(ownerRepo, sha, basePath, currentPath string) ([]File, error) {
	url := fmt.Sprintf("%s/repos/%s/contents/%s?ref=%s", f.baseURL, ownerRepo, currentPath, sha)
	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: unexpected status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response from %s: %w", url, err)
	}

	var entries []ghEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		// Single file response — try that shape
		var single ghEntry
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			return nil, fmt.Errorf("parse response from %s: %w", url, err)
		}
		entries = []ghEntry{single}
	}

	var files []File
	for _, entry := range entries {
		relPath := strings.TrimPrefix(entry.Path, basePath+"/")
		if relPath == entry.Path {
			relPath = entry.Path
		}

		switch entry.Type {
		case "file":
			content, err := f.download(entry.DownloadURL)
			if err != nil {
				return nil, err
			}
			files = append(files, File{Path: relPath, Content: content})
		case "dir":
			sub, err := f.walk(ownerRepo, sha, basePath, entry.Path)
			if err != nil {
				return nil, err
			}
			files = append(files, sub...)
		}
	}

	return files, nil
}

func (f *githubFetcher) download(url string) ([]byte, error) {
	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: unexpected status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read download from %s: %w", url, err)
	}
	return data, nil
}
