package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/iamy4n-dev/dpod-seed/internal/bundle"
	"github.com/iamy4n-dev/dpod-seed/internal/config"
	"github.com/iamy4n-dev/dpod-seed/internal/fetch"
	"github.com/iamy4n-dev/dpod-seed/internal/generator"
	"github.com/iamy4n-dev/dpod-seed/internal/hostconfig"
	"github.com/iamy4n-dev/dpod-seed/internal/registry"
	"github.com/iamy4n-dev/dpod-seed/internal/resolver"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactively select a distro and materialise DevPod config into this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		hc, _ := hostconfig.Load(hostconfig.DefaultPath())

		snap, _ := bundle.Load()

		var distros []generator.DistroRecord
		if snap != nil && len(snap.Distros) > 0 {
			distros = snap.Distros
		} else {
			// Fall back to live registry when running from a dev build.
			url := or(hc.RegistryURL, defaultRegistryURL)
			client := registry.NewClient(url, http.DefaultClient)
			entries, err := client.List()
			if err != nil {
				return fmt.Errorf("list distros: %w", err)
			}
			distros = registryEntriesToRecords(entries)
		}

		repos := resolver.RepoConfig{
			DistroRepo:       or(hc.Repos.Distro, defaultDistroRepo),
			DevcontainerRepo: or(hc.Repos.Devcontainer, defaultDevcontainerRepo),
			PackagesRepo:     or(hc.Repos.Packages, defaultPackagesRepo),
		}
		f := fetch.NewGitHubFetcher("", http.DefaultClient)
		r := resolver.NewResolver(f, repos)

		isTTY := term.IsTerminal(int(os.Stdin.Fd()))
		return runInit(os.Stdin, os.Stdout, isTTY, "dpod.yaml", "dpod.lock", ".", distros, r)
	},
}

// registryEntriesToRecords converts live registry entries to DistroRecord for
// the fallback network path. Devcontainer and package detail are empty because
// we haven't fetched the manifests; the review step will show what it can.
func registryEntriesToRecords(entries []registry.DistroEntry) []generator.DistroRecord {
	out := make([]generator.DistroRecord, len(entries))
	for i, e := range entries {
		out[i] = generator.DistroRecord{
			Name:        e.Name,
			Description: e.Description,
			LatestTag:   e.LatestTag,
			Status:      e.Status,
		}
	}
	return out
}

func runInit(r io.Reader, w io.Writer, isTTY bool, configPath, lockPath, baseDir string, distros []generator.DistroRecord, res resolver.Resolver) error {
	if !isTTY {
		return fmt.Errorf("no TTY detected: populate dpod.yaml manually, then run `dpod-seed sync`")
	}

	scanner := bufio.NewScanner(r)

	// Warn if dpod.yaml already exists.
	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprintf(w, "dpod.yaml already exists. Overwrite? [y/N] ")
		if !scanner.Scan() {
			return fmt.Errorf("unexpected end of input")
		}
		if !strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
			fmt.Fprintln(w, "Aborted.")
			return nil
		}
	}

	for {
		// Display numbered list.
		fmt.Fprintln(w, "\nAvailable distros:")
		for i, d := range distros {
			fmt.Fprintf(w, "  %d) %s — %s (%s)\n", i+1, d.Name, d.Description, d.LatestTag)
		}
		fmt.Fprintf(w, "Select a distro (1-%d): ", len(distros))

		var selected generator.DistroRecord
		for {
			if !scanner.Scan() {
				return fmt.Errorf("unexpected end of input")
			}
			n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err == nil && n >= 1 && n <= len(distros) {
				selected = distros[n-1]
				break
			}
			fmt.Fprintf(w, "Please enter a number between 1 and %d: ", len(distros))
		}

		printReview(w, selected)

		fmt.Fprint(w, "\n[y]es / [r]etry / [c]ancel: ")
		if !scanner.Scan() {
			return fmt.Errorf("unexpected end of input")
		}
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "y":
			cfg := &config.Config{Distro: selected.Name + "@" + selected.LatestTag}
			if err := config.Write(configPath, cfg); err != nil {
				return err
			}
			return runSync(w, configPath, lockPath, baseDir, res)
		case "r":
			continue
		default:
			fmt.Fprintln(w, "Cancelled.")
			return nil
		}
	}
}

// printReview prints the package-level review for a distro selection.
func printReview(w io.Writer, d generator.DistroRecord) {
	fmt.Fprintf(w, "\nReview — %s@%s\n", d.Name, d.LatestTag)
	if d.Devcontainer != "" {
		fmt.Fprintf(w, "  Devcontainer:  %s\n", d.Devcontainer)
	}
	if len(d.Packages) == 0 {
		fmt.Fprintln(w, "  Packages:      (none)")
		return
	}
	fmt.Fprintf(w, "  Packages (%d):", len(d.Packages))
	for i, p := range d.Packages {
		pin := p.Name
		if p.Version != "" {
			pin = p.Name + "@" + p.Version
		}
		if i == 0 {
			fmt.Fprintf(w, "  %s", pin)
		} else if i%3 == 0 {
			fmt.Fprintf(w, "\n               %s", pin)
		} else {
			fmt.Fprintf(w, ", %s", pin)
		}
	}
	fmt.Fprintln(w)
}
