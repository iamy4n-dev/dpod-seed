package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/duyanh-y4n/dpod-seed/internal/registry"
)

const defaultRegistryURL = "https://raw.githubusercontent.com/duyanh-y4n/distros/main/registry.yaml"

var registryURL string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available verified distros from the upstream registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := registry.NewClient(registryURL, http.DefaultClient)
		return runList(os.Stdout, client)
	},
}

func runList(w io.Writer, client registry.Client) error {
	entries, err := client.List()
	if err != nil {
		return fmt.Errorf("list distros: %w", err)
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION\tLATEST TAG")
	for _, e := range entries {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Name, e.Description, e.LatestTag)
	}
	return tw.Flush()
}

func init() {
	listCmd.Flags().StringVar(&registryURL, "registry", defaultRegistryURL, "URL of the registry.yaml file")
}
