package editor

import (
	"fmt"

	"ibinds/internal/cfgparse"
)

// Conflict reports that the target binding is already used by another control.
type Conflict struct {
	Control string
	Reason  string
}

// FindConflicts scans the file for controls whose current binding collides
// with the requested binding. The target control itself is ignored.
func FindConflicts(f *cfgparse.File, target string, b Binding) []Conflict {
	var out []Conflict
	for _, r := range f.Records {
		if r.Name == target {
			continue
		}
		switch {
		case b.HasKey && r.Kind == cfgparse.KindKey && r.Key == b.Key:
			out = append(out, Conflict{
				Control: r.Name,
				Reason:  fmt.Sprintf("key '%c' already bound to %s", b.Key, r.Name),
			})
		case b.HasButton && r.Kind == cfgparse.KindButton &&
			r.Button == b.Button && r.Device != nil && r.Device.InstanceGUID == b.DeviceInstanceGUID:
			out = append(out, Conflict{
				Control: r.Name,
				Reason:  fmt.Sprintf("button %d of device %s already bound to %s", b.Button, b.DeviceInstanceGUID, r.Name),
			})
		}
	}
	return out
}
