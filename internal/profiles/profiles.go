package profiles

import (
	"os"
	"path/filepath"
	"sort"
)

// Profile is a single iRacing control profile directory.
type Profile struct {
	ID          string
	Dir         string
	IsBaseline  bool
	HasControls bool
	HasCalib    bool
}

// Scan lists the valid profiles under dir (a directory of per-car folders).
// A folder is a profile when it contains controls.cfg or joyCalib.yaml.
// Baseline sorts first; the rest are alphabetical.
func Scan(dir string) ([]Profile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []Profile
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		d := filepath.Join(dir, id)
		hasControls := fileExists(filepath.Join(d, "controls.cfg"))
		hasCalib := fileExists(filepath.Join(d, "joyCalib.yaml"))
		if !hasControls && !hasCalib {
			continue
		}
		out = append(out, Profile{
			ID:          id,
			Dir:         d,
			IsBaseline:  id == "Baseline",
			HasControls: hasControls,
			HasCalib:    hasCalib,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].IsBaseline != out[j].IsBaseline {
			return out[i].IsBaseline
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
