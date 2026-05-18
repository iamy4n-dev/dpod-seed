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

	"github.com/duyanh-y4n/dpod-seed/internal/config"
	"github.com/duyanh-y4n/dpod-seed/internal/fetch"
	"github.com/duyanh-y4n/dpod-seed/internal/hostconfig"
	"github.com/duyanh-y4n/dpod-seed/internal/registry"
	"github.com/duyanh-y4n/dpod-seed/internal/resolver"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactively select a distro and materialise DevPod config into this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		hc, _ := hostconfig.Load(hostconfig.DefaultPath())
		url := or(hc.RegistryURL, defaultRegistryURL)
		client := registry.NewClient(url, http.DefaultClient)

		repos := resolver.RepoConfig{
			DistroRepo:       or(hc.Repos.Distro, defaultDistroRepo),
			DevcontainerRepo: or(hc.Repos.Devcontainer, defaultDevcontainerRepo),
			PackagesRepo:     or(hc.Repos.Packages, defaultPackagesRepo),
		}
		f := fetch.NewGitHubFetcher("", http.DefaultClient)
		r := resolver.NewResolver(f, repos)

		isTTY := term.IsTerminal(int(os.Stdin.Fd()))
		return runInit(os.Stdin, os.Stdout, isTTY, "dpod.yaml", "dpod.lock", ".", client, r)
	},
}

func runInit(r io.Reader, w io.Writer, isTTY bool, configPath, lockPath, baseDir string, client registry.Client, res resolver.Resolver) error {
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

	distros, err := client.List()
	if err != nil {
		return fmt.Errorf("list distros: %w", err)
	}

	for {
		// Display numbered list.
		fmt.Fprintln(w, "\nAvailable distros:")
		for i, d := range distros {
			fmt.Fprintf(w, "  %d) %s — %s (%s)\n", i+1, d.Name, d.Description, d.LatestTag)
		}
		fmt.Fprintf(w, "Select a distro (1-%d): ", len(distros))

		var selected registry.DistroEntry
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

		// Resolve to show review.
		entries, err := res.Resolve(selected.Name, selected.LatestTag, config.Overrides{})
		if err != nil {
			return fmt.Errorf("resolve distro: %w", err)
		}

		fmt.Fprintf(w, "\nReview — %s@%s will write %d file(s):\n", selected.Name, selected.LatestTag, len(entries))
		for _, e := range entries {
			fmt.Fprintf(w, "  %s\n", e.DestPath)
		}

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
