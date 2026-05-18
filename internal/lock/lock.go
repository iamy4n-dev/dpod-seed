package lock

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type File struct {
	Path string `yaml:"path"`
	Repo string `yaml:"repo"`
	SHA  string `yaml:"sha"`
}

type Lock struct {
	Files []File `yaml:"files"`
}

type Diff struct {
	Added   []File
	Updated []File
	Removed []File
}

func Read(path string) (*Lock, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Lock{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var l Lock
	if err := yaml.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &l, nil
}

func Write(path string, l *Lock) error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func Compute(old *Lock, newFiles []File) Diff {
	oldByPath := make(map[string]File, len(old.Files))
	for _, f := range old.Files {
		oldByPath[f.Path] = f
	}

	newByPath := make(map[string]struct{}, len(newFiles))
	var d Diff

	for _, f := range newFiles {
		newByPath[f.Path] = struct{}{}
		if existing, ok := oldByPath[f.Path]; !ok {
			d.Added = append(d.Added, f)
		} else if existing.SHA != f.SHA {
			d.Updated = append(d.Updated, f)
		}
	}

	for _, f := range old.Files {
		if _, ok := newByPath[f.Path]; !ok {
			d.Removed = append(d.Removed, f)
		}
	}

	return d
}
