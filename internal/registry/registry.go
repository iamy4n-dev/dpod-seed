package registry

import (
	"fmt"
	"io"
	"net/http"

	"gopkg.in/yaml.v3"
)

type DistroEntry struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	LatestTag    string `yaml:"latestTag"`
	ChangelogURL string `yaml:"changelogUrl"`
	Status       string `yaml:"status"`
}

type registryPayload struct {
	Distros []DistroEntry `yaml:"distros"`
}

type Client interface {
	List() ([]DistroEntry, error)
}

type registryClient struct {
	url        string
	httpClient *http.Client
}

func NewClient(url string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &registryClient{url: url, httpClient: httpClient}
}

func (c *registryClient) List() ([]DistroEntry, error) {
	resp, err := c.httpClient.Get(c.url)
	if err != nil {
		return nil, fmt.Errorf("fetch registry %s: %w", c.url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch registry %s: unexpected status %d", c.url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read registry response: %w", err)
	}

	var payload registryPayload
	if err := yaml.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}

	return payload.Distros, nil
}
