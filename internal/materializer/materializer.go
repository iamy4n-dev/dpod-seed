package materializer

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iamy4n-dev/dpod-seed/internal/lock"
	"github.com/iamy4n-dev/dpod-seed/internal/patchapplier"
	"github.com/iamy4n-dev/dpod-seed/internal/resolver"
)

// Summary reports how many files were written and removed.
type Summary struct {
	Written int
	Removed int
}

// Materialize writes entries to disk under baseDir, applies patch files, and deletes removed files.
// patchPaths are relative to baseDir. Missing files in removes are silently skipped.
func Materialize(baseDir string, entries []resolver.ManifestEntry, removes []lock.File, patchPaths []string) (Summary, error) {
	var s Summary

	for _, e := range entries {
		dest := filepath.Join(baseDir, e.DestPath)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return s, fmt.Errorf("create dir for %s: %w", e.DestPath, err)
		}
		if err := os.WriteFile(dest, e.Content, 0o644); err != nil {
			return s, fmt.Errorf("write %s: %w", e.DestPath, err)
		}
		s.Written++
	}

	for _, patchPath := range patchPaths {
		if err := applyPatchFile(baseDir, patchPath); err != nil {
			return s, fmt.Errorf("apply patch %s: %w", patchPath, err)
		}
	}

	for _, f := range removes {
		dest := filepath.Join(baseDir, f.Path)
		if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
			return s, fmt.Errorf("remove %s: %w", f.Path, err)
		}
		s.Removed++
	}

	return s, nil
}

// applyPatchFile reads a unified diff patch and applies it to the file named in its "+++ b/" header.
func applyPatchFile(baseDir, patchPath string) error {
	patchContent, err := os.ReadFile(filepath.Join(baseDir, patchPath))
	if err != nil {
		return fmt.Errorf("read patch: %w", err)
	}

	targetPath, err := patchTarget(patchContent)
	if err != nil {
		return err
	}

	dest := filepath.Join(baseDir, targetPath)
	src, err := os.ReadFile(dest)
	if err != nil {
		return fmt.Errorf("read patch target %s: %w", targetPath, err)
	}

	patched, err := patchapplier.Apply(src, patchContent)
	if err != nil {
		return fmt.Errorf("apply to %s: %w", targetPath, err)
	}

	return os.WriteFile(dest, patched, 0o644)
}

// patchTarget extracts the target file path from a "+++ b/<path>" line.
func patchTarget(patch []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(patch))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+++ b/") {
			return strings.TrimPrefix(line, "+++ b/"), nil
		}
		if strings.HasPrefix(line, "+++ ") {
			return strings.TrimPrefix(line, "+++ "), nil
		}
	}
	return "", fmt.Errorf("patch: no '+++ b/' header found")
}
