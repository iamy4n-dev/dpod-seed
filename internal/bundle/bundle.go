package bundle

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/iamy4n-dev/dpod-seed/internal/generator"
)

//go:embed snapshot.json
var snapshotBytes []byte

// Load returns the embedded snapshot. Returns an empty Output (no distros)
// if the snapshot is the stub placeholder, so callers can fall back to network.
func Load() (*generator.Output, error) {
	var out generator.Output
	if err := json.Unmarshal(snapshotBytes, &out); err != nil {
		return nil, fmt.Errorf("parse embedded bundle: %w", err)
	}
	return &out, nil
}
