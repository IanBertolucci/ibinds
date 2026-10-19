package ops

import (
	"os"
	"path/filepath"
	"testing"

	"ibinds/internal/editor"
)

func setupProfile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	pd := filepath.Join(dir, "f4")
	if err := os.MkdirAll(pd, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "cfgparse", "testdata", "f4_key_u.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pd, "controls.cfg"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestPlanBindingDiffShowsOnlyTargetBytes(t *testing.T) {
	profDir := setupProfile(t)

	res, err := PlanBinding(profDir, "f4", "TractionControlArm", editor.Binding{Key: 'I', HasKey: true}, false)
	if err != nil {
		t.Fatalf("PlanBinding: %v", err)
	}
	// 'U' (0x55) -> 'I' (0x49) changes exactly one byte at the key offset 0xbd0.
	if res.Diff != "1 byte changed at 0xbd0" {
		t.Fatalf("Diff = %q, want %q", res.Diff, "1 byte changed at 0xbd0")
	}
}
