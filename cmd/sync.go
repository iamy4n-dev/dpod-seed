package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/duyanh-y4n/dpod-seed/internal/hostconfig"
	"github.com/duyanh-y4n/dpod-seed/internal/resolver"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Re-materialise DevPod config from dpod.yaml, overwriting owned files",
	RunE: func(cmd *cobra.Command, args []string) error {
		hc, _ := hostconfig.Load(hostconfig.DefaultPath())
		_ = resolver.RepoConfig{
			DistroRepo:       or(hc.Repos.Distro, defaultDistroRepo),
			DevcontainerRepo: or(hc.Repos.Devcontainer, defaultDevcontainerRepo),
			PackagesRepo:     or(hc.Repos.Packages, defaultPackagesRepo),
		}
		fmt.Fprintln(os.Stderr, "not implemented")
		os.Exit(1)
		return nil
	},
}
