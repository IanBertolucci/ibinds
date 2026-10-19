package ops

import (
	"fmt"
	"os"
	"path/filepath"

	"ibinds/internal/calib"
	"ibinds/internal/cfgparse"
	"ibinds/internal/editor"
	"ibinds/internal/fileops"
)

// PlanResult is the outcome of planning a binding change against one profile.
type PlanResult struct {
	Conflicts []editor.Conflict
	Warning   string
	Diff      string
	Force     bool
}

// PlanBinding computes, without writing, what applying the binding to the
// profile would do: conflicts, a missing-device warning, and a diff summary.
func PlanBinding(profDir, profileID, control string, binding editor.Binding, force bool) (PlanResult, error) {
	cfgPath := filepath.Join(profDir, profileID, "controls.cfg")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return PlanResult{}, err
	}
	f, err := cfgparse.Parse(data)
	if err != nil {
		return PlanResult{}, err
	}

	conflicts := editor.FindConflicts(f, control, binding)

	warning := ""
	if binding.HasButton {
		has, err := profileHasDevice(profDir, profileID, binding.DeviceInstanceGUID)
		if err != nil {
			return PlanResult{}, err
		}
		if !has {
			warning = fmt.Sprintf("device %s is not listed in this profile's joyCalib.yaml", binding.DeviceInstanceGUID)
		}
	}

	newData, _, err := editor.ApplyBinding(f, control, binding)
	if err != nil {
		return PlanResult{}, err
	}
	diff := fileops.DiffSummary(fileops.Diff(data, newData))

	return PlanResult{Conflicts: conflicts, Warning: warning, Diff: diff, Force: force}, nil
}

// profileHasDevice reports whether the profile's joyCalib.yaml lists the
// given InstanceGUID.
func profileHasDevice(profDir, profileID, instanceGUID string) (bool, error) {
	calibPath := filepath.Join(profDir, profileID, "joyCalib.yaml")
	data, err := os.ReadFile(calibPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	c, err := calib.Parse(data)
	if err != nil {
		return false, err
	}
	for _, d := range c.Devices {
		if d.InstanceGUID == instanceGUID {
			return true, nil
		}
	}
	return false, nil
}
