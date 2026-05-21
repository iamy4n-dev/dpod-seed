package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate directory layout for a new distro or package bundle",
}

var scaffoldDistroCmd = &cobra.Command{
	Use:   "distro <name>",
	Short: "Generate a new distro directory layout with a distro.yaml template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScaffoldDistro(args[0], ".")
	},
}

var scaffoldPackageCmd = &cobra.Command{
	Use:   "package <name>",
	Short: "Generate a new package bundle directory layout with a manifest.yaml template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScaffoldPackage(args[0], ".")
	},
}

func runScaffoldDistro(name, baseDir string) error {
	dir := filepath.Join(baseDir, "distros", name)
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("path %q already exists", filepath.Join("distros", name))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	distroYAML := `# distro.yaml — defines a sealed, versioned DevPod preset.
#
# Pin the devcontainer profile (from the devcontainer/ repo):
#   devcontainer: arch-base@v2.0.0
#
# Pin package bundles (from the packages/ repo):
#   packages:
#     - shell-zsh@v1.3.0
#     - k8s-tools@v1.1.0
name: ` + name + `
description: ""
devcontainer: arch-base@v0.1.0
packages: []
`
	readme := "# " + name + "\n\nDescribe this distro.\n"

	if err := os.WriteFile(filepath.Join(dir, "distro.yaml"), []byte(distroYAML), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644); err != nil {
		return err
	}
	fmt.Printf("Created distros/%s/\n", name)
	return nil
}

func runScaffoldPackage(name, baseDir string) error {
	dir := filepath.Join(baseDir, "packages", name)
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("path %q already exists", filepath.Join("packages", name))
	}
	dotfilesDir := filepath.Join(dir, "dotfiles")
	if err := os.MkdirAll(dotfilesDir, 0o755); err != nil {
		return err
	}

	manifestYAML := `# manifest.yaml — optional placement overrides for this bundle.
#
# By default, files under dotfiles/ map to .devcontainer/ preserving path.
# Use this file only for exceptions to that convention.
#
# Example:
#   files:
#     - src: dotfiles/.special
#       dest: .devcontainer/special-location
files: []
`
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(manifestYAML), 0o644); err != nil {
		return err
	}
	fmt.Printf("Created packages/%s/\n", name)
	return nil
}

func init() {
	scaffoldCmd.AddCommand(scaffoldDistroCmd, scaffoldPackageCmd)
}
