package cmd

const (
	defaultDistroRepo       = "github.com/duyanh-y4n/distros"
	defaultDevcontainerRepo = "github.com/duyanh-y4n/devcontainer"
	defaultPackagesRepo     = "github.com/duyanh-y4n/packages"
)

// or returns override if non-empty, otherwise fallback.
func or(override, fallback string) string {
	if override != "" {
		return override
	}
	return fallback
}
