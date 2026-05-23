package validate

import (
	"fmt"
	"net/http"
	"strings"
)

type githubRefChecker struct {
	baseURL    string
	httpClient *http.Client
}

// NewGitHubRefChecker returns a RefChecker that verifies tags via the GitHub API.
func NewGitHubRefChecker(baseURL string, httpClient *http.Client) RefChecker {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &githubRefChecker{baseURL: baseURL, httpClient: httpClient}
}

func (c *githubRefChecker) TagExists(repo, tag string) (bool, error) {
	ownerRepo := strings.TrimPrefix(repo, "github.com/")
	url := fmt.Sprintf("%s/repos/%s/git/ref/refs/tags/%s", c.baseURL, ownerRepo, tag)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("check tag %s in %s: %w", tag, repo, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("check tag %s in %s: unexpected status %d", tag, repo, resp.StatusCode)
	}
}
