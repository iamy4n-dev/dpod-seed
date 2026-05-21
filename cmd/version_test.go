package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion_printsVersion(t *testing.T) {
	var out bytes.Buffer
	runVersion(&out)
	got := out.String()
	if !strings.HasPrefix(got, "dpod-seed ") {
		t.Errorf("output should start with %q, got: %q", "dpod-seed ", got)
	}
	if !strings.Contains(got, version) {
		t.Errorf("output should contain version %q, got: %q", version, got)
	}
}
