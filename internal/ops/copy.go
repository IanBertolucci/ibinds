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

	// ErrSourceControls marks a missing, unreadable, or unparsable source
	// controls.cfg. Callers should treat it as a client error.
	ErrSourceControls = errors.New("ops: source controls.cfg unavailable")
)

// CopyPlan is a planned copy plus the target's exact bytes before and after,
// so callers can write without re-reading controls.cfg.
type CopyPlan struct {
	PlanResult
	Old []byte
	New []byte
}

// CopySource is a parsed source profile reused for one or more targets.
type CopySource struct {
	file *cfgparse.File
}

// LoadCopySource reads and parses the source profile's controls.cfg once.
func LoadCopySource(profDir, sourceID string) (*CopySource, error) {
	path := filepath.Join(profDir, sourceID, "controls.cfg")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceControls, err)
	}
	f, err := cfgparse.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceControls, err)
	}
	return &CopySource{file: f}, nil
}

// PlanCopy computes, without writing, what copying the bindings of `controls`
// from the source profile to the target profile would do. It reuses the
// existing PlanResult shape; Warning joins the per-control warnings.
func PlanCopy(profDir, sourceID, targetID string, controls []string, force bool) (PlanResult, error) {
	src, err := LoadCopySource(profDir, sourceID)
	if err != nil {
		return PlanResult{}, err
	}
	res, err := PlanCopyFromSource(src, profDir, targetID, controls, force)
	if err != nil {
		return PlanResult{}, err
	}
	return res.PlanResult, nil
}

// PlanCopyFromSource computes a copy against an already loaded source.
func PlanCopyFromSource(src *CopySource, profDir, targetID string, controls []string, force bool) (CopyPlan, error) {
	orig, newData, conflicts, warnings, err := copyBindings(src.file, profDir, targetID, controls)
	if err != nil {
		return CopyPlan{}, err
	}
	return CopyPlan{
		PlanResult: PlanResult{
			Conflicts: conflicts,
			Warning:   strings.Join(warnings, "; "),
			Diff:      fileops.DiffSummary(fileops.Diff(orig, newData)),
			Force:     force,
		},
		Old: orig,
		New: newData,
	}, nil
}

// ApplyCopy returns the target profile's controls.cfg bytes after applying
// the copyable controls' bindings from the source profile.
func ApplyCopy(profDir, sourceID, targetID string, controls []string) ([]byte, error) {
	src, err := LoadCopySource(profDir, sourceID)
	if err != nil {
		return nil, err
	}
	res, err := PlanCopyFromSource(src, profDir, targetID, controls, false)
	if err != nil {
		return nil, err
	}
	return res.New, nil
}

// copyBindings derives each control's source binding and applies them to the
// target in sequence. It returns the target's original and new bytes, the
// accumulated conflicts, and the per-control warnings.
func copyBindings(src *cfgparse.File, profDir, targetID string, controls []string) (orig, new []byte, conflicts []editor.Conflict, warnings []string, err error) {
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
		rec, serr := sourceRecord(src, control)
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
