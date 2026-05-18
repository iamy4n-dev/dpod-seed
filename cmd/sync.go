package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/duyanh-y4n/dpod-seed/internal/config"
	"github.com/duyanh-y4n/dpod-seed/internal/fetch"
	"github.com/duyanh-y4n/dpod-seed/internal/hostconfig"
	"github.com/duyanh-y4n/dpod-seed/internal/lock"
	"github.com/duyanh-y4n/dpod-seed/internal/materializer"
	"github.com/duyanh-y4n/dpod-seed/internal/resolver"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Re-materialise DevPod config from dpod.yaml, overwriting owned files",
	RunE: func(cmd *cobra.Command, args []string) error {
		hc, _ := hostconfig.Load(hostconfig.DefaultPath())
		repos := resolver.RepoConfig{
			DistroRepo:       or(hc.Repos.Distro, defaultDistroRepo),
			DevcontainerRepo: or(hc.Repos.Devcontainer, defaultDevcontainerRepo),
			PackagesRepo:     or(hc.Repos.Packages, defaultPackagesRepo),
		}
		f := fetch.NewGitHubFetcher("", http.DefaultClient)
		r := resolver.NewResolver(f, repos)
		return runSync(os.Stdout, "dpod.yaml", "dpod.lock", ".", r)
	},
}

func runSync(w io.Writer, configPath, lockPath, baseDir string, res resolver.Resolver) error {
	cfg, err := config.Read(configPath)
	if err != nil {
		return fmt.Errorf("read dpod.yaml: %w", err)
	}

	distroName, tag := splitDistroPin(cfg.Distro)
	entries, err := res.Resolve(distroName, tag, cfg.Overrides)
	if err != nil {
		return err
	}

	existing, err := lock.Read(lockPath)
	if err != nil {
		return err
	}

	newFiles := toLockedFiles(entries)
	diff := lock.Compute(existing, newFiles)

	toWriteSet := make(map[string]bool, len(diff.Added)+len(diff.Updated))
	for _, f := range diff.Added {
		toWriteSet[f.Path] = true
	}
	for _, f := range diff.Updated {
		toWriteSet[f.Path] = true
	}
	var toWrite []resolver.ManifestEntry
	for _, e := range entries {
		if toWriteSet[e.DestPath] {
			toWrite = append(toWrite, e)
		}
	}

	if _, err := materializer.Materialize(baseDir, toWrite, diff.Removed, cfg.Overrides.Patches); err != nil {
		return err
	}

	if err := lock.Write(lockPath, &lock.Lock{Files: newFiles}); err != nil {
		return err
	}

	fmt.Fprintf(w, "%d added, %d updated, %d removed\n", len(diff.Added), len(diff.Updated), len(diff.Removed))
	return nil
}

func toLockedFiles(entries []resolver.ManifestEntry) []lock.File {
	files := make([]lock.File, len(entries))
	for i, e := range entries {
		files[i] = lock.File{Path: e.DestPath, Repo: e.SrcRepo, SHA: e.SHA}
	}
	return files
}
