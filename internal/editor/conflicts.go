package editor

import (
	"fmt"
	"strings"

	"ibinds/internal/cfgparse"
)

// Conflict reports that the target binding is already used by another control.
type Conflict struct {
	Control string `json:"control"`
	Reason  string `json:"reason"`
}

// replayOrCameraPrefixes are the control-name prefixes for replay and camera
// controls. iRacing keeps these in a separate namespace from driving
// controls, so sharing a key/button with them is not a conflict.
var replayOrCameraPrefixes = []string{"Cam", "DCam", "Rpy", "NextRpyCam", "PrevRpyCam"}

// isReplayOrCamera reports whether a control belongs to the replay/camera
// namespace (safe to share a binding with a driving control).
func isReplayOrCamera(name string) bool {
	for _, p := range replayOrCameraPrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// FindConflicts scans the file for controls whose current binding collides
// with the requested binding. The target control itself is ignored, as are
// replay/camera controls (a separate namespace).
func FindConflicts(f *cfgparse.File, target string, b Binding) []Conflict {
	out := make([]Conflict, 0)
	for _, r := range f.Records {
		if r.Name == target {
			continue
		}
		if isReplayOrCamera(r.Name) {
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
