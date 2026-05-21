package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the dpod-seed version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runVersion(os.Stdout)
	},
}

func runVersion(out io.Writer) {
	fmt.Fprintf(out, "dpod-seed %s\n", version)
}
