package editor

import (
	"bytes"
	"errors"
	"testing"

	"ibinds/internal/cfgparse"
)

func TestCopyBindingKeyReproducesRealFixture(t *testing.T) {
	f := loadFile(t, "f4_btn128.cfg")
	src := findRecord(loadFile(t, "f4_key_u.cfg"), "TractionControlArm")
	if src == nil {
		t.Fatal("TractionControlArm not found in source")
	}
	out, rec, err := CopyBinding(f, "TractionControlArm", *src)
	if err != nil {
		t.Fatalf("CopyBinding: %v", err)
	}
	if rec.Kind != cfgparse.KindKey || rec.Key != 'U' {
		t.Fatalf("record = %+v, want KindKey Key='U'", rec)
	}
	if want := loadBytes(t, "f4_key_u.cfg"); !bytes.Equal(out, want) {
		t.Fatalf("output not byte-identical to f4_key_u.cfg")
	}
}

func TestCopyBindingButtonReproducesRealFixture(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	src := findRecord(loadFile(t, "f4_btn130.cfg"), "TractionControlArm")
	if src == nil {
		t.Fatal("TractionControlArm not found in source")
	}
	out, rec, err := CopyBinding(f, "TractionControlArm", *src)
	if err != nil {
		t.Fatalf("CopyBinding: %v", err)
	}
	if rec.Kind != cfgparse.KindButton || rec.Button != 130 {
		t.Fatalf("record = %+v, want KindButton Button=130", rec)
	}
	if want := loadBytes(t, "f4_btn130.cfg"); !bytes.Equal(out, want) {
		t.Fatalf("output not byte-identical to f4_btn130.cfg")
	}
}

func TestCopyBindingMultiButton(t *testing.T) {
	srcFile := loadFile(t, "f4_key_u.cfg")
	src := findRecord(srcFile, "TearOffVisor")
	if src == nil {
		t.Fatal("TearOffVisor not found in source")
	}
	if len(src.Buttons) < 2 {
		t.Fatalf("TearOffVisor.Buttons = %v, want at least 2", src.Buttons)
	}

	// Make a target whose TearOffVisor is a single button (Btn 9), then copy
	// the multi-button source over it.
	target := loadFile(t, "f4_btn128.cfg")
	single, _, err := ApplyBinding(target, "TearOffVisor", Binding{
		Button: 9, HasButton: true,
		DeviceInstanceGUID: src.Device.InstanceGUID,
		ProductGUID:        src.Device.ProductGUID,
	})
	if err != nil {
		t.Fatalf("ApplyBinding: %v", err)
	}
	targetFile, err := cfgparse.Parse(single)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out, rec, err := CopyBinding(targetFile, "TearOffVisor", *src)
	if err != nil {
		t.Fatalf("CopyBinding: %v", err)
	}
	if !slicesEqual(rec.Buttons, src.Buttons) {
		t.Fatalf("record.Buttons = %v, want %v", rec.Buttons, src.Buttons)
	}

	// The mask region of the copied record must be byte-identical to the source's.
	got, err := cfgparse.Parse(out)
	if err != nil {
		t.Fatalf("Parse out: %v", err)
	}
	copied := findRecord(got, "TearOffVisor")
	if copied == nil {
		t.Fatal("TearOffVisor not found in output")
	}
	srcMask := src.Opaque[len(src.Opaque)-cfgparse.FixedRegionSize:]
	copiedMask := copied.Opaque[len(copied.Opaque)-cfgparse.FixedRegionSize:]
	if !bytes.Equal(srcMask[cfgparse.TypeOffset:], copiedMask[cfgparse.TypeOffset:]) {
		t.Fatalf("binding region not byte-identical to source")
	}
}

func TestCopyBindingPreservesFlagsAndCategory(t *testing.T) {
	target := loadFile(t, "f4_btn128.cfg")
	src := findRecord(loadFile(t, "f4_key_u.cfg"), "TractionControlArm")
	if src == nil {
		t.Fatal("TractionControlArm not found")
	}
	targetRec := findRecord(target, "TractionControlArm")
	if targetRec == nil {
		t.Fatal("TractionControlArm not found in target")
	}
	beforeFlags := target.Raw[targetRec.PayloadStart : targetRec.PayloadStart+cfgparse.TypeOffset]

	out, _, err := CopyBinding(target, "TractionControlArm", *src)
	if err != nil {
		t.Fatalf("CopyBinding: %v", err)
	}
	afterFlags := out[targetRec.PayloadStart : targetRec.PayloadStart+cfgparse.TypeOffset]
	if !bytes.Equal(beforeFlags, afterFlags) {
		t.Fatalf("flags/category changed: % x -> % x", beforeFlags, afterFlags)
	}
}

func TestCopyBindingUnknownControl(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	src := findRecord(f, "TractionControlArm")
	_, _, err := CopyBinding(f, "NoSuchControl", *src)
	if !errors.Is(err, ErrUnknownControl) {
		t.Fatalf("err = %v, want ErrUnknownControl", err)
	}
}

func TestCopyBindingRefusesAxisTarget(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	src := findRecord(f, "TractionControlArm")
	_, _, err := CopyBinding(f, "SteerLeft", *src)
	if !errors.Is(err, ErrUnsupportedKind) {
		t.Fatalf("err = %v, want ErrUnsupportedKind", err)
	}
}

func TestCopyBindingRefusesNonCopyableSource(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	unbound := findRecord(f, "Ignition")
	if unbound == nil {
		t.Fatal("Ignition not found")
	}
	_, _, err := CopyBinding(f, "TractionControlArm", *unbound)
	if !errors.Is(err, ErrNotCopyable) {
		t.Fatalf("err = %v, want ErrNotCopyable", err)
	}
}

func TestFindCopyConflictsMultiButton(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	tearoff := findRecord(f, "TearOffVisor")
	if tearoff == nil {
		t.Fatal("TearOffVisor not found")
	}
	if len(tearoff.Buttons) < 2 {
		t.Fatalf("TearOffVisor.Buttons = %v, want at least 2", tearoff.Buttons)
	}
	// TearOffVisor shares button 33 with ToggleWindshieldWipers and HeadlightFlash.
	conflicts := FindCopyConflicts(f, "PushToPass", *tearoff)
	if !containsControl(conflicts, "ToggleWindshieldWipers") {
		t.Fatalf("conflicts = %+v, want ToggleWindshieldWipers", conflicts)
	}
	if !containsControl(conflicts, "HeadlightFlash") {
		t.Fatalf("conflicts = %+v, want HeadlightFlash", conflicts)
	}
}

func TestFindCopyConflictsKey(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	src := findRecord(f, "TractionControlArm")
	if src == nil {
		t.Fatal("TractionControlArm not found")
	}
	conflicts := FindCopyConflicts(f, "Ignition", *src)
	if len(conflicts) != 1 {
		t.Fatalf("len(conflicts) = %d, want 1 (%+v)", len(conflicts), conflicts)
	}
	if !containsControl(conflicts, "TractionControlArm") {
		t.Fatalf("conflicts = %+v, want TractionControlArm", conflicts)
	}
}

func slicesEqual(a, b []int) bool {
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
