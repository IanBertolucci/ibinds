package fileops

import (
	"bytes"
	"testing"
)

func TestDiffFindsSingleByteChange(t *testing.T) {
	orig := make([]byte, 4096)
	next := make([]byte, 4096)
	orig[0xbd0] = 0x55
	next[0xbd0] = 0x49

	hunks := Diff(orig, next)
	if len(hunks) != 1 {
		t.Fatalf("len(hunks) = %d, want 1", len(hunks))
	}
	h := hunks[0]
	if h.Offset != 0xbd0 {
		t.Fatalf("Offset = %#x, want 0xbd0", h.Offset)
	}
	if !bytes.Equal(h.Orig, []byte{0x55}) || !bytes.Equal(h.New, []byte{0x49}) {
		t.Fatalf("hunk = %+v, want Orig=[0x55] New=[0x49]", h)
	}
}

func TestDiffIdentical(t *testing.T) {
	data := []byte("hello")
	hunks := Diff(data, data)
	if len(hunks) != 0 {
		t.Fatalf("len(hunks) = %d, want 0", len(hunks))
	}
	if DiffSummary(hunks) != "no changes" {
		t.Fatalf("DiffSummary = %q, want %q", DiffSummary(hunks), "no changes")
	}
}

func TestDiffSummaryFormat(t *testing.T) {
	h := []Hunk{{Offset: 0xbd0, Orig: []byte{0x55}, New: []byte{0x49}}}
	got := DiffSummary(h)
	if got != "1 byte changed at 0xbd0" {
		t.Fatalf("DiffSummary = %q, want %q", got, "1 byte changed at 0xbd0")
	}
}
