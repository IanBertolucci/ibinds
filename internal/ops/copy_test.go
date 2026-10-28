package ops

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ibinds/internal/cfgparse"
	"ibinds/internal/editor"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "cfgparse", "testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func writeProfile(t *testing.T, profDir, id, cfgName string, edits map[string]editor.Binding) string {
	t.Helper()
	pd := filepath.Join(profDir, id)
	if err := os.MkdirAll(pd, 0o755); err != nil {
		t.Fatal(err)
	}
	data := readFixture(t, cfgName)
	if err := os.WriteFile(filepath.Join(pd, "controls.cfg"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	calib, err := os.ReadFile(filepath.Join("..", "calib", "testdata", "f4_joycalib.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pd, "joyCalib.yaml"), calib, 0o644); err != nil {
		t.Fatal(err)
	}
	if len(edits) > 0 {
		f, err := cfgparse.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		cur := data
		for control, b := range edits {
			f, err = cfgparse.Parse(cur)
			if err != nil {
				t.Fatal(err)
			}
			cur, _, err = editor.ApplyBinding(f, control, b)
			if err != nil {
				t.Fatalf("ApplyBinding %s: %v", control, err)
			}
		}
		if err := os.WriteFile(filepath.Join(pd, "controls.cfg"), cur, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return pd
}

func parse(t *testing.T, data []byte) *cfgparse.File {
	t.Helper()
	f, err := cfgparse.Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return f
}

func TestSourceRecordKey(t *testing.T) {
	src := parse(t, readFixture(t, "f4_key_u.cfg"))
	rec, err := sourceRecord(src, "TractionControlArm")
	if err != nil {
		t.Fatalf("sourceRecord: %v", err)
	}
	if rec.Kind != cfgparse.KindKey || rec.Key != 'U' {
		t.Fatalf("record = %+v, want KindKey Key='U'", rec)
	}
}

func TestSourceRecordButton(t *testing.T) {
	src := parse(t, readFixture(t, "f4_btn130.cfg"))
	rec, err := sourceRecord(src, "TractionControlArm")
	if err != nil {
		t.Fatalf("sourceRecord: %v", err)
	}
	if rec.Kind != cfgparse.KindButton || !slicesEq(rec.Buttons, []int{130}) {
		t.Fatalf("record = %+v, want KindButton Buttons=[130]", rec)
	}
}

func TestSourceRecordMultiButton(t *testing.T) {
	src := parse(t, readFixture(t, "f4_key_u.cfg"))
	rec, err := sourceRecord(src, "TearOffVisor")
	if err != nil {
		t.Fatalf("sourceRecord: %v", err)
	}
	if !slicesEq(rec.Buttons, []int{7, 33}) {
		t.Fatalf("record.Buttons = %v, want [7 33]", rec.Buttons)
	}
}

func TestSourceRecordUnbound(t *testing.T) {
	src := parse(t, readFixture(t, "f4_key_u.cfg"))
	if _, err := sourceRecord(src, "Ignition"); !errors.Is(err, errNotCopyable) {
		t.Fatalf("err = %v, want errNotCopyable", err)
	}
}

func TestSourceRecordMissing(t *testing.T) {
	src := parse(t, readFixture(t, "f4_key_u.cfg"))
	if _, err := sourceRecord(src, "Nope"); !errors.Is(err, errControlMissing) {
		t.Fatalf("err = %v, want errControlMissing", err)
	}
}

func TestPlanCopyKeyDiff(t *testing.T) {
	profDir := t.TempDir()
	writeProfile(t, profDir, "src", "f4_key_u.cfg", nil)
	writeProfile(t, profDir, "dst", "f4_btn128.cfg", nil)

	res, err := PlanCopy(profDir, "src", "dst", []string{"TractionControlArm"}, false)
	if err != nil {
		t.Fatalf("PlanCopy: %v", err)
	}
	if len(res.Conflicts) != 0 {
		t.Fatalf("conflicts = %+v, want none", res.Conflicts)
	}
	if res.Warning != "" {
		t.Fatalf("warning = %q, want empty", res.Warning)
	}
	if res.Diff == "no changes" {
		t.Fatalf("Diff = %q, want actual changes", res.Diff)
	}
}

func TestApplyCopyKeyReproducesRealFixture(t *testing.T) {
	profDir := t.TempDir()
	writeProfile(t, profDir, "src", "f4_key_u.cfg", nil)
	writeProfile(t, profDir, "dst", "f4_btn128.cfg", nil)

	out, err := ApplyCopy(profDir, "src", "dst", []string{"TractionControlArm"})
	if err != nil {
		t.Fatalf("ApplyCopy: %v", err)
	}
	if want := readFixture(t, "f4_key_u.cfg"); !bytes.Equal(out, want) {
		t.Fatalf("output not byte-identical to f4_key_u.cfg")
	}
}

func TestApplyCopyButtonReproducesRealFixture(t *testing.T) {
	profDir := t.TempDir()
	writeProfile(t, profDir, "src", "f4_btn130.cfg", nil)
	writeProfile(t, profDir, "dst", "f4_key_u.cfg", nil)

	out, err := ApplyCopy(profDir, "src", "dst", []string{"TractionControlArm"})
	if err != nil {
		t.Fatalf("ApplyCopy: %v", err)
	}
	if want := readFixture(t, "f4_btn130.cfg"); !bytes.Equal(out, want) {
		t.Fatalf("output not byte-identical to f4_btn130.cfg")
	}
}

func TestApplyCopyMultiButtonReproducesRealFixture(t *testing.T) {
	profDir := t.TempDir()
	writeProfile(t, profDir, "src", "f4_key_u.cfg", nil)
	// Clear the target's TearOffVisor to a single button so the copy visibly
	// restores the multi-button mask.
	writeProfile(t, profDir, "dst", "f4_btn128.cfg", map[string]editor.Binding{
		"TearOffVisor": {
			Button: 9, HasButton: true,
			DeviceInstanceGUID: "{991805D0-2EB6-11F1-800A-444553540000}",
			ProductGUID:        "{0002346E-0000-0000-0000-504944564944}",
		},
	})

	out, err := ApplyCopy(profDir, "src", "dst", []string{"TearOffVisor"})
	if err != nil {
		t.Fatalf("ApplyCopy: %v", err)
	}
	got := parse(t, out)
	var rec *cfgparse.Record
	for i := range got.Records {
		if got.Records[i].Name == "TearOffVisor" {
			rec = &got.Records[i]
			break
		}
	}
	if rec == nil {
		t.Fatal("TearOffVisor not found")
	}
	if !slicesEq(rec.Buttons, []int{7, 33}) {
		t.Fatalf("TearOffVisor.Buttons = %v, want [7 33]", rec.Buttons)
	}
}

func TestPlanCopyConflict(t *testing.T) {
	profDir := t.TempDir()
	writeProfile(t, profDir, "src", "f4_key_u.cfg", map[string]editor.Binding{
		"Ignition": {Key: 'U', HasKey: true},
	})
	writeProfile(t, profDir, "dst", "f4_key_u.cfg", nil)

	res, err := PlanCopy(profDir, "src", "dst", []string{"Ignition"}, false)
	if err != nil {
		t.Fatalf("PlanCopy: %v", err)
	}
	if len(res.Conflicts) != 1 || res.Conflicts[0].Control != "TractionControlArm" {
		t.Fatalf("conflicts = %+v, want exactly TractionControlArm", res.Conflicts)
	}
}

func TestPlanCopyMultiButtonConflict(t *testing.T) {
	profDir := t.TempDir()
	writeProfile(t, profDir, "src", "f4_key_u.cfg", nil)
	writeProfile(t, profDir, "dst", "f4_btn128.cfg", map[string]editor.Binding{
		"PushToPass": {
			Button: 7, HasButton: true,
			DeviceInstanceGUID: "{991805D0-2EB6-11F1-800A-444553540000}",
			ProductGUID:        "{0002346E-0000-0000-0000-504944564944}",
		},
	})

	res, err := PlanCopy(profDir, "src", "dst", []string{"TearOffVisor"}, false)
	if err != nil {
		t.Fatalf("PlanCopy: %v", err)
	}
	found := false
	for _, c := range res.Conflicts {
		if c.Control == "PushToPass" {
			found = true
		}
	}
	if !found {
		t.Fatalf("conflicts = %+v, want PushToPass", res.Conflicts)
	}
}

func TestPlanCopyWarnsDeviceMissing(t *testing.T) {
	profDir := t.TempDir()
	writeProfile(t, profDir, "src", "f4_btn130.cfg", map[string]editor.Binding{
		"Ignition": {
			Button: 130, HasButton: true,
			DeviceInstanceGUID: "{11111111-2222-3333-4444-555555555555}",
			ProductGUID:        "{0002346E-0000-0000-0000-504944564944}",
		},
	})
	writeProfile(t, profDir, "dst", "f4_key_u.cfg", nil)

	res, err := PlanCopy(profDir, "src", "dst", []string{"Ignition"}, false)
	if err != nil {
		t.Fatalf("PlanCopy: %v", err)
	}
	if res.Warning == "" {
		t.Fatalf("warning = %q, want non-empty (device missing)", res.Warning)
	}
}

func TestApplyCopyOldFormatTargetPreservesTrailingBytes(t *testing.T) {
	profDir := t.TempDir()
	writeProfile(t, profDir, "src", "f4_key_u.cfg", nil)
	// The old-format file has 2 trailing bytes after its record region; the
	// copy must reproduce them byte-for-byte.
	writeProfile(t, profDir, "dst", "audigto_oldfmt.cfg", nil)

	out, err := ApplyCopy(profDir, "src", "dst", []string{"TractionControlArm"})
	if err != nil {
		t.Fatalf("ApplyCopy: %v", err)
	}
	want := readFixture(t, "audigto_oldfmt.cfg")
	if len(out) != len(want) {
		t.Fatalf("len(out) = %d, want %d", len(out), len(want))
	}
	if out[len(out)-1] != 0x20 || out[len(out)-2] != 0x20 {
		t.Fatalf("trailing bytes = % x, want 20 20", out[len(out)-2:])
	}
	// The only changed bytes are the TractionControlArm record; everything
	// else (including the trailing bytes) is identical to the source target.
	got := parse(t, out)
	if len(got.Records) != 367 {
		t.Fatalf("len(Records) = %d, want 367", len(got.Records))
	}
}

func slicesEq(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
