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
	// 'U' is bound to TractionControlArm (driving) and to two camera controls
	// (CamFocusDepthDec, CamFNumberDec) — the camera controls are a separate
	// namespace and must not be flagged.
	conflicts := FindConflicts(f, "Ignition", Binding{Key: 'U', HasKey: true})
	if len(conflicts) != 1 {
		t.Fatalf("len(conflicts) = %d, want 1 (%+v)", len(conflicts), conflicts)
	}
	if !containsControl(conflicts, "TractionControlArm") {
		t.Fatalf("conflicts missing TractionControlArm: %+v", conflicts)
	}
}

func TestDetectIgnoresReplayControls(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	conflicts := FindConflicts(f, "Ignition", Binding{Key: 'U', HasKey: true})
	for _, c := range conflicts {
		if isReplayOrCamera(c.Control) {
			t.Fatalf("replay/camera control %q was flagged as a conflict", c.Control)
		}
	}
}

func TestDetectCameraTargetDoesNotConflictDriving(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	// CamFNumberDec is a camera control bound to 'U'. Targeting a camera
	// control must not flag the driving control TractionControlArm (separate
	// namespace).
	conflicts := FindConflicts(f, "CamFNumberDec", Binding{Key: 'U', HasKey: true})
	if containsControl(conflicts, "TractionControlArm") {
		t.Fatalf("driving control flagged for a camera target: %+v", conflicts)
	}
}

func TestDetectButtonConflictMultiButton(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	tearoff := findRecord(f, "TearOffVisor")
	if tearoff == nil {
		t.Fatal("TearOffVisor not found")
	}
	if len(tearoff.Buttons) < 2 {
		t.Fatalf("TearOffVisor.Buttons = %v, want at least 2", tearoff.Buttons)
	}
	second := tearoff.Buttons[1]
	conflicts := FindConflicts(f, "Ignition", Binding{
		Button:             second,
		HasButton:          true,
		DeviceInstanceGUID: tearoff.Device.InstanceGUID,
		ProductGUID:        tearoff.Device.ProductGUID,
	})
	if !containsControl(conflicts, "TearOffVisor") {
		t.Fatalf("conflicts = %+v, want TearOffVisor (button %d)", conflicts, second)
	}
}

func TestDetectNoConflictWhenUnbound(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	conflicts := FindConflicts(f, "Ignition", Binding{Key: 'q', HasKey: true})
	if len(conflicts) != 0 {
		t.Fatalf("len(conflicts) = %d, want 0 (%+v)", len(conflicts), conflicts)
	}
}

func TestFindConflictsReturnsEmptySlice(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	conflicts := FindConflicts(f, "Ignition", Binding{Key: 'q', HasKey: true})
	if conflicts == nil {
		t.Fatal("FindConflicts returned nil, want an empty slice (marshals to [] not null)")
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
