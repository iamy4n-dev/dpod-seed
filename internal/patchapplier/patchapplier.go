package patchapplier

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

type hunkLine struct {
	kind byte // ' ', '+', '-'
	text string
}

type hunk struct {
	srcStart int // 1-indexed line in source where this hunk begins
	ops      []hunkLine
}

// Apply applies a unified diff patch to src and returns the patched result.
// Returns an error if the patch is malformed or cannot be applied to src.
func Apply(src, patch []byte) ([]byte, error) {
	hunks, err := parseHunks(patch)
	if err != nil {
		return nil, err
	}

	srcLines := strings.Split(string(src), "\n")
	var out []string
	srcPos := 0 // 0-indexed position in srcLines

	for _, h := range hunks {
		startIdx := h.srcStart - 1 // convert to 0-indexed
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx < srcPos {
			return nil, fmt.Errorf("patch: overlapping or out-of-order hunks")
		}
		// copy unchanged lines before this hunk
		out = append(out, srcLines[srcPos:startIdx]...)
		srcPos = startIdx

		for _, op := range h.ops {
			switch op.kind {
			case ' ':
				if srcPos >= len(srcLines) {
					return nil, fmt.Errorf("patch: context line beyond end of source at line %d", srcPos+1)
				}
				if srcLines[srcPos] != op.text {
					return nil, fmt.Errorf("patch: context mismatch at line %d: want %q, got %q", srcPos+1, op.text, srcLines[srcPos])
				}
				out = append(out, srcLines[srcPos])
				srcPos++
			case '-':
				if srcPos >= len(srcLines) {
					return nil, fmt.Errorf("patch: removal beyond end of source at line %d", srcPos+1)
				}
				if srcLines[srcPos] != op.text {
					return nil, fmt.Errorf("patch: cannot apply removal at line %d: want %q, got %q — context mismatch", srcPos+1, op.text, srcLines[srcPos])
				}
				srcPos++ // consume without emitting
			case '+':
				out = append(out, op.text)
			}
		}
	}

	// append any remaining source lines after the last hunk
	out = append(out, srcLines[srcPos:]...)
	return []byte(strings.Join(out, "\n")), nil
}

func parseHunks(patch []byte) ([]hunk, error) {
	lines := strings.Split(string(patch), "\n")
	var hunks []hunk
	var cur *hunk

	for _, l := range lines {
		// hunk header
		if strings.HasPrefix(l, "@@") {
			m := hunkHeaderRe.FindStringSubmatch(l)
			if m == nil {
				return nil, fmt.Errorf("patch: malformed hunk header: %q", l)
			}
			if cur != nil {
				hunks = append(hunks, *cur)
			}
			start, _ := strconv.Atoi(m[1])
			cur = &hunk{srcStart: start}
			continue
		}
		// skip diff header lines and empty lines
		if cur == nil || l == "" || strings.HasPrefix(l, "---") || strings.HasPrefix(l, "+++") || strings.HasPrefix(l, "\\") {
			continue
		}
		switch l[0] {
		case ' ', '+', '-':
			cur.ops = append(cur.ops, hunkLine{l[0], l[1:]})
		default:
			return nil, fmt.Errorf("patch: malformed line: %q", l)
		}
	}
	if cur != nil {
		hunks = append(hunks, *cur)
	}
	return hunks, nil
}
