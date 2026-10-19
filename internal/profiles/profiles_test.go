package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanListsBaselineAndCars(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"Baseline", "f4", "audigto", "junk"} {
		if err := os.MkdirAll(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	touch := func(rel string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, rel), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	touch("Baseline/controls.cfg")
	touch("f4/controls.cfg")
	touch("f4/joyCalib.yaml")
	touch("audigto/joyCalib.yaml")

	got, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3 (profiles: %+v)", len(got), got)
	}

	byID := map[string]Profile{}
	for _, p := range got {
		byID[p.ID] = p
	}

	if f4, ok := byID["f4"]; !ok {
		t.Fatalf("missing f4 profile")
	} else if !f4.HasControls || !f4.HasCalib {
		t.Fatalf("f4 = %+v, want HasControls=true HasCalib=true", f4)
	}

	if audigto, ok := byID["audigto"]; !ok {
		t.Fatalf("missing audigto profile")
	} else if audigto.HasControls || !audigto.HasCalib {
		t.Fatalf("audigto = %+v, want HasControls=false HasCalib=true", audigto)
	}

	if _, ok := byID["junk"]; ok {
		t.Fatalf("junk (empty dir) should not be a profile")
	}

	if baseline, ok := byID["Baseline"]; !ok {
		t.Fatalf("missing Baseline profile")
	} else if !baseline.IsBaseline {
		t.Fatalf("Baseline = %+v, want IsBaseline=true", baseline)
	}
}

func TestScanSortsBaselineFirst(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"Baseline", "zzz"} {
		if err := os.MkdirAll(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "Baseline", "controls.cfg"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "zzz", "controls.cfg"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "Baseline" {
		t.Fatalf("first = %q, want Baseline (order: %v)", got[0].ID, got)
	}
}
