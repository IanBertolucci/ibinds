package fileops

import (
	"fmt"
	"strings"
)

// Hunk is a contiguous run of differing bytes between two files.
type Hunk struct {
	Offset int
	Orig   []byte
	New    []byte
}

// Diff returns the contiguous runs of differing bytes. It is empty when the
// inputs are identical. Bytes beyond the shorter input are reported as one
// trailing hunk.
func Diff(orig, next []byte) []Hunk {
	n := len(orig)
	if len(next) < n {
		n = len(next)
	}
	var hunks []Hunk
	i := 0
	for i < n {
		if orig[i] == next[i] {
			i++
			continue
		}
		start := i
		for i < n && orig[i] != next[i] {
			i++
		}
		hunks = append(hunks, Hunk{Offset: start, Orig: orig[start:i], New: next[start:i]})
	}
	if len(orig) != len(next) {
		hunks = append(hunks, Hunk{Offset: n, Orig: orig[n:], New: next[n:]})
	}
	return hunks
}

// DiffSummary renders a human-readable description of the hunks.
func DiffSummary(h []Hunk) string {
	if len(h) == 0 {
		return "no changes"
	}
	parts := make([]string, 0, len(h))
	for _, hk := range h {
		n := len(hk.Orig)
		if len(hk.New) > n {
			n = len(hk.New)
		}
		unit := "byte"
		if n != 1 {
			unit = "bytes"
		}
		parts = append(parts, fmt.Sprintf("%d %s changed at 0x%x", n, unit, hk.Offset))
	}
	return strings.Join(parts, ", ")
}
