package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Overrides struct {
	Packages PackageOverrides `yaml:"packages,omitempty"`
	Patches  []string         `yaml:"patches,omitempty"`
}

type PackageOverrides struct {
	Add    []string `yaml:"add,omitempty"`
	Remove []string `yaml:"remove,omitempty"`
}

type Config struct {
	Distro    string    `yaml:"distro"`
	Overrides Overrides `yaml:"overrides,omitempty"`
}

func Read(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Distro == "" {
		return nil, fmt.Errorf("parse %s: missing required field 'distro'", path)
	}
	return &cfg, nil
}

func Write(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
