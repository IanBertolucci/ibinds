package editor

import (
	"testing"

	"ibinds/internal/cfgparse"
)

func containsControl(cs []Conflict, name string) bool {
	for _, c := range cs {
		if c.Control == name {
			return true
		}
	}
	return false
}

func TestDetectKeyConflict(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	// 'U' is bound to TractionControlArm, CamFocusDepthDec and CamFNumberDec
	// in the real fixture (the profile already has duplicate key bindings).
	conflicts := FindConflicts(f, "Ignition", Binding{Key: 'U', HasKey: true})
	if len(conflicts) != 3 {
		t.Fatalf("len(conflicts) = %d, want 3 (%+v)", len(conflicts), conflicts)
	}
	if !containsControl(conflicts, "TractionControlArm") {
		t.Fatalf("conflicts missing TractionControlArm: %+v", conflicts)
	}
}

func TestDetectNoConflictWhenUnbound(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	conflicts := FindConflicts(f, "Ignition", Binding{Key: 'q', HasKey: true})
	if len(conflicts) != 0 {
		t.Fatalf("len(conflicts) = %d, want 0 (%+v)", len(conflicts), conflicts)
	}
}

func TestDetectButtonConflictSameDevice(t *testing.T) {
	f := loadFile(t, "f4_btn128.cfg")
	out, _, err := ApplyBinding(f, "Ignition", Binding{
		Button: 128, HasButton: true,
		DeviceInstanceGUID: mozaGUID,
		ProductGUID:        mozaProdGUID,
	})
	if err != nil {
		t.Fatalf("ApplyBinding: %v", err)
	}
	f2, err := cfgparse.Parse(out)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	conflicts := FindConflicts(f2, "TractionControlArm", Binding{
		Button: 128, HasButton: true,
		DeviceInstanceGUID: mozaGUID,
		ProductGUID:        mozaProdGUID,
	})
	if len(conflicts) != 1 || conflicts[0].Control != "Ignition" {
		t.Fatalf("conflicts = %+v, want exactly Ignition", conflicts)
	}
}

func TestDetectButtonConflictDifferentDeviceIsNotConflict(t *testing.T) {
	f := loadFile(t, "f4_btn128.cfg")
	out, _, err := ApplyBinding(f, "Ignition", Binding{
		Button: 128, HasButton: true,
		DeviceInstanceGUID: "{11111111-2222-3333-4444-555555555555}",
		ProductGUID:        mozaProdGUID,
	})
	if err != nil {
		t.Fatalf("ApplyBinding: %v", err)
	}
	f2, err := cfgparse.Parse(out)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	conflicts := FindConflicts(f2, "TractionControlArm", Binding{
		Button: 128, HasButton: true,
		DeviceInstanceGUID: mozaGUID,
		ProductGUID:        mozaProdGUID,
	})
	if len(conflicts) != 0 {
		t.Fatalf("conflicts = %+v, want 0 (different device)", conflicts)
	}
}

func TestDetectIgnoresTarget(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	// CamToggleCamCtrlActive is the only control bound to 'Z', so targeting it
	// must not report a conflict.
	conflicts := FindConflicts(f, "CamToggleCamCtrlActive", Binding{Key: 'Z', HasKey: true})
	if len(conflicts) != 0 {
		t.Fatalf("conflicts = %+v, want 0 (target ignored)", conflicts)
	}
}
