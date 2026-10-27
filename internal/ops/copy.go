package ops

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ibinds/internal/cfgparse"
	"ibinds/internal/editor"
	"ibinds/internal/fileops"
)

var (
	errControlMissing = errors.New("ops: control not found in source profile")
	errNotCopyable    = errors.New("ops: control binding is not copyable")
)

// PlanCopy computes, without writing, what copying the bindings of `controls`
// from the source profile to the target profile would do. It reuses the
// existing PlanResult shape; Warning joins the per-control warnings.
func PlanCopy(profDir, sourceID, targetID string, controls []string, force bool) (PlanResult, error) {
	orig, newData, conflicts, warnings, err := copyBindings(profDir, sourceID, targetID, controls)
	if err != nil {
		return PlanResult{}, err
	}
	return PlanResult{
		Conflicts: conflicts,
		Warning:   strings.Join(warnings, "; "),
		Diff:      fileops.DiffSummary(fileops.Diff(orig, newData)),
		Force:     force,
	}, nil
}

// ApplyCopy returns the target profile's controls.cfg bytes after applying
// the copyable controls' bindings from the source profile.
func ApplyCopy(profDir, sourceID, targetID string, controls []string) ([]byte, error) {
	_, newData, _, _, err := copyBindings(profDir, sourceID, targetID, controls)
	if err != nil {
		return nil, err
	}
	return newData, nil
}

// copyBindings reads the source and target, derives each control's source
// binding, and applies them in sequence. It returns the target's original and
// new bytes, the accumulated conflicts, and the per-control warnings.
func copyBindings(profDir, sourceID, targetID string, controls []string) (orig, new []byte, conflicts []editor.Conflict, warnings []string, err error) {
	srcPath := filepath.Join(profDir, sourceID, "controls.cfg")
	srcData, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	srcFile, err := cfgparse.Parse(srcData)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	tgtPath := filepath.Join(profDir, targetID, "controls.cfg")
	tgtData, err := os.ReadFile(tgtPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	orig = tgtData

	cur := tgtData
	curFile, err := cfgparse.Parse(cur)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	conflicts = make([]editor.Conflict, 0)
	warnings = make([]string, 0)

	for _, control := range controls {
		rec, serr := sourceRecord(srcFile, control)
		if serr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", control, serr))
			continue
		}
		conflicts = append(conflicts, editor.FindCopyConflicts(curFile, control, rec)...)

		if rec.Kind == cfgparse.KindButton {
			has, herr := profileHasDevice(profDir, targetID, rec.Device.InstanceGUID)
			if herr != nil {
				return nil, nil, nil, nil, herr
			}
			if !has {
				warnings = append(warnings, fmt.Sprintf("device %s is not listed in this profile's joyCalib.yaml", rec.Device.InstanceGUID))
			}
		}

		newData, _, aerr := editor.CopyBinding(curFile, control, rec)
		if aerr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", control, aerr))
			continue
		}
		cur = newData
		curFile, err = cfgparse.Parse(cur)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	return orig, cur, conflicts, warnings, nil
}

// sourceRecord returns the source profile's record for a control when its
// binding is copyable (key or button).
func sourceRecord(src *cfgparse.File, control string) (cfgparse.Record, error) {
	for _, r := range src.Records {
		if r.Name != control {
			continue
		}
		switch r.Kind {
		case cfgparse.KindKey, cfgparse.KindButton:
			return r, nil
		default:
			return cfgparse.Record{}, errNotCopyable
		}
	}
	return cfgparse.Record{}, errControlMissing
}
