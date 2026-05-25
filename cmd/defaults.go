package cmd

import (
	"net/http"
	"os"
)

const (
	defaultDistroRepo       = "github.com/iamy4n-dev/distros"
	defaultDevcontainerRepo = "github.com/iamy4n-dev/devcontainer"
	defaultPackagesRepo     = "github.com/iamy4n-dev/packages"
)

// or returns override if non-empty, otherwise fallback.
func or(override, fallback string) string {
	if override != "" {
		return override
	}
	return fallback
}

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(r)
}

// authedHTTPClient returns an HTTP client that sends GITHUB_TOKEN if set,
// falling back to the default unauthenticated client.
func authedHTTPClient() *http.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return http.DefaultClient
	}
	return &http.Client{Transport: &bearerTransport{token: token, base: http.DefaultTransport}}
}
