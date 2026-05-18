package hostconfig

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type RepoOverrides struct {
	Devcontainer string `yaml:"devcontainer"`
	Packages     string `yaml:"packages"`
	Distro       string `yaml:"distro"`
}

type Config struct {
	RegistryURL string        `yaml:"registryUrl"`
	Repos       RepoOverrides `yaml:"repos"`
}

// Load reads config from path. Returns zero-value Config if the file does not exist.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// DefaultPath returns the canonical path for the user-level config file.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dpod-seed", "config.yaml")
}
