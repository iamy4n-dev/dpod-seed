package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/iamy4n-dev/dpod-seed/internal/fetch"
	"github.com/iamy4n-dev/dpod-seed/internal/generator"
	"github.com/iamy4n-dev/dpod-seed/internal/hostconfig"
	"github.com/iamy4n-dev/dpod-seed/internal/registry"
)

const (
	defaultRegistryURL = "https://raw.githubusercontent.com/iamy4n-dev/distros/main/registry.yaml"
	defaultDistroRepo  = "github.com/iamy4n-dev/distros"
	defaultOutputPath  = "docs/src/data/registry-data.json"
)

type generateConfig struct {
	registryURL string
	distroRepo  string
	githubBase  string
	outputPath  string
	httpClient  *http.Client
}

type authTransport struct {
	token     string
	transport http.RoundTripper
}

func (t *authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+t.token)
	return t.transport.RoundTrip(r)
}

func newAuthClient(token string, base http.RoundTripper) *http.Client {
	if base == nil {
		base = http.DefaultTransport
	}
	return &http.Client{Transport: &authTransport{token: token, transport: base}}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	hc, _ := hostconfig.Load(hostconfig.DefaultPath())
	token := os.Getenv("GITHUB_TOKEN")

	var httpClient *http.Client
	if token != "" {
		httpClient = newAuthClient(token, nil)
	} else {
		httpClient = http.DefaultClient
	}

	cfg := generateConfig{
		registryURL: firstNonEmpty(hc.RegistryURL, defaultRegistryURL),
		distroRepo:  firstNonEmpty(hc.Repos.Distro, defaultDistroRepo),
		githubBase:  "",
		outputPath:  defaultOutputPath,
		httpClient:  httpClient,
	}
	return runGenerate(cfg)
}

func runGenerate(cfg generateConfig) error {
	reg := registry.NewClient(cfg.registryURL, cfg.httpClient)
	f := fetch.NewGitHubFetcher(cfg.githubBase, cfg.httpClient)

	out, err := os.Create(cfg.outputPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", cfg.outputPath, err)
	}
	defer out.Close()

	if err := generator.Generate(reg, f, cfg.distroRepo, out); err != nil {
		os.Remove(cfg.outputPath)
		return err
	}

	fmt.Fprintf(os.Stderr, "wrote %s\n", cfg.outputPath)
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
