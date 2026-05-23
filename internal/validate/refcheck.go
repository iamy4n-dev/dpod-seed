package validate

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// RefChecker verifies that a version tag exists in a remote repository.
type RefChecker interface {
	TagExists(repo, tag string) (bool, error)
}

// Options controls reference resolution in DistroDirWithOptions.
// A nil Checker disables reference checks (offline mode).
type Options struct {
	DevcontainerRepo string
	PackagesRepo     string
	Checker          RefChecker
}

// DistroDirWithOptions validates a distro directory and, when opts.Checker is
// non-nil, verifies devcontainer and package version pins against their remote repos.
func DistroDirWithOptions(dir string, opts Options) []string {
	errs := DistroDir(dir)
	if opts.Checker == nil {
		return errs
	}

	distroPath := dir + "/distro.yaml"
	d, err := parseDistroYAML(distroPath)
	if err != nil {
		return errs
	}

	if d.Devcontainer != "" {
		pin := d.Devcontainer
		tag := tagFromPin(pin)
		exists, err := opts.Checker.TagExists(opts.DevcontainerRepo, tag)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: devcontainer pin %q: %v", distroPath, pin, err))
		} else if !exists {
			errs = append(errs, fmt.Sprintf("%s: devcontainer pin %q not found in %s", distroPath, pin, opts.DevcontainerRepo))
		}
	}

	for _, pkg := range d.Packages {
		pin := pkg
		tag := tagFromPin(pin)
		exists, err := opts.Checker.TagExists(opts.PackagesRepo, tag)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: package pin %q: %v", distroPath, pin, err))
		} else if !exists {
			errs = append(errs, fmt.Sprintf("%s: package pin %q not found in %s", distroPath, pin, opts.PackagesRepo))
		}
	}

	return errs
}

func tagFromPin(pin string) string {
	if i := strings.LastIndex(pin, "@"); i >= 0 {
		return pin[i+1:]
	}
	return pin
}

func parseDistroYAML(path string) (*distroYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d distroYAML
	if err := yaml.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}
