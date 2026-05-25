package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iamy4n-dev/dpod-seed/internal/config"
	"github.com/iamy4n-dev/dpod-seed/internal/fetch"
	"github.com/iamy4n-dev/dpod-seed/internal/hostconfig"
	"github.com/iamy4n-dev/dpod-seed/internal/resolver"
	"github.com/iamy4n-dev/dpod-seed/internal/validate"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate dpod.yaml or a distros/packages content repository",
}

var validateProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Resolve dpod.yaml without writing files — suitable for CI",
	RunE: func(cmd *cobra.Command, args []string) error {
		hc, _ := hostconfig.Load(hostconfig.DefaultPath())
		repos := resolver.RepoConfig{
			DistroRepo:       or(hc.Repos.Distro, defaultDistroRepo),
			DevcontainerRepo: or(hc.Repos.Devcontainer, defaultDevcontainerRepo),
			PackagesRepo:     or(hc.Repos.Packages, defaultPackagesRepo),
		}
		f := fetch.NewGitHubFetcher("", http.DefaultClient)
		r := resolver.NewResolver(f, repos)
		return runValidate("dpod.yaml", os.Stdout, r)
	},
}

func runValidate(configPath string, out io.Writer, r resolver.Resolver) error {
	cfg, err := config.Read(configPath)
	if err != nil {
		return fmt.Errorf("read dpod.yaml: %w", err)
	}

	distroName, tag := splitDistroPin(cfg.Distro)
	entries, err := r.Resolve(distroName, tag, cfg.Overrides)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "OK — resolved %d files\n", len(entries))
	return nil
}

var validateRepoOffline bool

var validateRepoCmd = &cobra.Command{
	Use:   "repo [path]",
	Short: "Validate all distros/ or packages/ content in a repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}
		hc, _ := hostconfig.Load(hostconfig.DefaultPath())
		opts := validate.Options{
			DevcontainerRepo: or(hc.Repos.Devcontainer, defaultDevcontainerRepo),
			PackagesRepo:     or(hc.Repos.Packages, defaultPackagesRepo),
		}
		if !validateRepoOffline {
			opts.Checker = validate.NewGitHubRefChecker("", authedHTTPClient())
		}
		return runValidateRepo(root, os.Stdout, opts)
	},
}

func runValidateRepo(root string, out io.Writer, opts validate.Options) error {
	errs := validate.RunWithOptions(root, opts)
	for _, e := range errs {
		fmt.Fprintln(out, e)
	}
	if len(errs) > 0 {
		return fmt.Errorf("validation failed with %d error(s)", len(errs))
	}
	fmt.Fprintln(out, "OK — all content valid")
	return nil
}

// splitDistroPin splits "name@tag" from dpod.yaml into (name, tag).
func splitDistroPin(pin string) (name, tag string) {
	if i := strings.LastIndex(pin, "@"); i >= 0 {
		return pin[:i], pin[i+1:]
	}
	return pin, ""
}

func init() {
	validateCmd.AddCommand(validateProjectCmd, validateRepoCmd)
	validateRepoCmd.Flags().BoolVar(&validateRepoOffline, "offline", false, "skip cross-repo reference resolution")
}
