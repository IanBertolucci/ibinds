package editor

import (
	"errors"

	"ibinds/internal/cfgparse"
)

// ErrNotCopyable indicates a source record has no binding that can be copied
// (it is unbound, an axis, or an unknown kind).
var ErrNotCopyable = errors.New("editor: source record has no copyable binding")

// CopyBinding overwrites the target control's binding with the source
// record's, copying the binding region byte-for-byte (type + key/mask +
// reserved + GUID + ProductID + reserved + PIDVID, fixed-region offsets +8
// through +67). flags (+0) and category (+4) are left untouched. This makes
// multi-button bindings copy faithfully: the whole button mask is transferred.
func CopyBinding(f *cfgparse.File, control string, src cfgparse.Record) ([]byte, cfgparse.Record, error) {
	if !cfgparse.ValidMagic(f.Raw) {
		return nil, cfgparse.Record{}, cfgparse.ErrBadMagic
	}

	var rec *cfgparse.Record
	for i := range f.Records {
		if f.Records[i].Name == control {
			rec = &f.Records[i]
			break
		}
	}
	if rec == nil {
		return nil, cfgparse.Record{}, ErrUnknownControl
	}
	if rec.Kind == cfgparse.KindAxis || rec.Kind == cfgparse.KindUnknown {
		return nil, cfgparse.Record{}, ErrUnsupportedKind
	}
	if src.Kind != cfgparse.KindKey && src.Kind != cfgparse.KindButton {
		return nil, cfgparse.Record{}, ErrNotCopyable
	}

	out := make([]byte, len(f.Raw))
	copy(out, f.Raw)

	srcFixed := src.Opaque[len(src.Opaque)-cfgparse.FixedRegionSize:]
	copy(out[rec.PayloadStart+cfgparse.TypeOffset:rec.PayloadStart+cfgparse.FixedRegionSize],
		srcFixed[cfgparse.TypeOffset:])

	newFile, err := cfgparse.Parse(out)
	if err != nil {
		return nil, cfgparse.Record{}, err
	}
	for i := range newFile.Records {
		if newFile.Records[i].Name == control {
			return out, newFile.Records[i], nil
		}
	}
	return nil, cfgparse.Record{}, ErrUnknownControl
}

// FindCopyConflicts scans the file for controls whose current binding collides
// with the source record's binding, using the same namespace scoping as
// FindConflicts. A key source conflicts with another key control sharing the
// key; a button source conflicts when any of its buttons overlaps another
// button control's buttons on the same device.
func FindCopyConflicts(f *cfgparse.File, target string, src cfgparse.Record) []Conflict {
	out := make([]Conflict, 0)
	for _, r := range f.Records {
		if r.Name == target {
			continue
		}
		if isReplayOrCamera(r.Name) != isReplayOrCamera(target) {
			continue
		}
		switch src.Kind {
		case cfgparse.KindKey:
			if r.Kind == cfgparse.KindKey && r.Key == src.Key {
				out = append(out, Conflict{
					Control: r.Name,
					Reason:  "key '" + string([]byte{src.Key}) + "' already bound to " + r.Name,
				})
			}
		case cfgparse.KindButton:
			if r.Kind == cfgparse.KindButton &&
				r.Device != nil && src.Device != nil &&
				r.Device.InstanceGUID == src.Device.InstanceGUID &&
				buttonsOverlap(r.Buttons, src.Buttons) {
				out = append(out, Conflict{
					Control: r.Name,
					Reason:  "button overlap of device " + src.Device.InstanceGUID + " already bound to " + r.Name,
				})
			}
		}
	}
	return out
}

// buttonsOverlap reports whether the two sorted button slices share any value.
func buttonsOverlap(a, b []int) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			return true
		}
		if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return false
}
