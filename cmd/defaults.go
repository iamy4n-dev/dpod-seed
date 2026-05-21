package cmd

const (
	defaultDistroRepo       = "github.com/iamy4n-dev/distros"
	defaultDevcontainerRepo = "github.com/iamy4n-dev/devcontainer"
	defaultPackagesRepo     = "github.com/iamy4n-dev/packages"
)

// or returns override if non-empty, otherwise fallback.
func or(override, fallback string) string {
	if override != "" {
		return override
	}
	return fallback
}
