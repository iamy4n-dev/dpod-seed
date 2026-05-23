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
	outputPath         = "docs/src/data/registry-data.json"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	hc, _ := hostconfig.Load(hostconfig.DefaultPath())
	registryURL := firstNonEmpty(hc.RegistryURL, defaultRegistryURL)
	distroRepo := firstNonEmpty(hc.Repos.Distro, defaultDistroRepo)

	reg := registry.NewClient(registryURL, http.DefaultClient)
	f := fetch.NewGitHubFetcher("", http.DefaultClient)

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outputPath, err)
	}
	defer out.Close()

	if err := generator.Generate(reg, f, distroRepo, out); err != nil {
		os.Remove(outputPath)
		return err
	}

	fmt.Fprintf(os.Stderr, "wrote %s\n", outputPath)
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
