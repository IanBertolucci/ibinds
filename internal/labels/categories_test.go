package labels

import (
	"os"
	"path/filepath"
	"testing"

	"ibinds/internal/cfgparse"
)

func TestControlGroupKnown(t *testing.T) {
	cases := []struct {
		id  string
		cat string
		sub string
	}{
		{"Throttle", "Cockpit", ""},
		{"RecenterTiltAxis", "Driver's View", ""},
		{"Pause", "Interface", ""},
		{"BlackBoxF1", "Black Box", ""},
		{"AutoChat8", "Chat", "Quick Text Chat"},
		{"CamFOVInc", "Camera & Screen Capture", ""},
		{"BrakeBiasInc", "Car Setup Adjustments", "Brakes"},
	}
	for _, tc := range cases {
		g := ControlGroupFor(tc.id)
		if g.Category != tc.cat || g.Subcategory != tc.sub {
			t.Errorf("ControlGroupFor(%q) = {%q, %q}, want {%q, %q}",
				tc.id, g.Category, g.Subcategory, tc.cat, tc.sub)
		}
	}
}

func TestControlGroupOrder(t *testing.T) {
	ids := []string{
		"Throttle",
		"RecenterTiltAxis",
		"BlackBoxF1",
		"RpyToStart",
		"BrakeBiasInc",
	}
	prev := -1
	for _, id := range ids {
		g := ControlGroupFor(id)
		if g.Order <= prev {
			t.Errorf("ControlGroupFor(%q).Order = %d, want > %d", id, g.Order, prev)
		}
		prev = g.Order
	}
}

func TestControlGroupFallback(t *testing.T) {
	g := ControlGroupFor("NotARealControl")
	if g.Category != "Other" {
		t.Fatalf("fallback category = %q, want Other", g.Category)
	}
}

func TestEveryFixtureControlHasGroup(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "cfgparse", "testdata", "*.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no cfgparse testdata fixtures found")
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		f, err := cfgparse.Parse(data)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, rec := range f.Records {
			g := ControlGroupFor(rec.Name)
			if g.Category == "" || g.Category == "Other" {
				t.Errorf("control %q has no in-game category", rec.Name)
			}
		}
	}
}
