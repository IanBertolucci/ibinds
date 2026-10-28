package labels

import (
	"os"
	"path/filepath"
	"testing"

	"ibinds/internal/cfgparse"
)

func TestCarFallback(t *testing.T) {
	if got := Car("audigto"); got != "audigto" {
		t.Fatalf("Car(audigto) = %q, want audigto", got)
	}
	if got := Car("totallyunknowncar"); got != "totallyunknowncar" {
		t.Fatalf("Car(unknown) = %q, want the id itself", got)
	}
}

func TestControlFallback(t *testing.T) {
	for _, id := range []string{"NotARealControl", "Throttle", "Gear1"} {
		if got := Control(id); got != id {
			t.Fatalf("Control(%q) = %q, want the id itself", id, got)
		}
	}
}

func TestKnownMappings(t *testing.T) {
	if got := Car("f4"); got != "F4" {
		t.Fatalf("Car(f4) = %q, want F4", got)
	}
	if got := Control("LowFuelAccept"); got != "Low Fuel Accept" {
		t.Fatalf("Control(LowFuelAccept) = %q, want \"Low Fuel Accept\"", got)
	}
}

func TestControlGroupedLabels(t *testing.T) {
	cases := []struct {
		id   string
		want string
	}{
		{"BrakeBiasInc", "Brake Bias Set ↑"},
		{"BrakeBiasDec", "Brake Bias Set ↓"},
		{"BrakeBiasLevel", "Brake Bias Set"},
		{"ABSInc", "ABS Set ↑"},
		{"ABSDec", "ABS Set ↓"},
		{"TractionControl2Inc", "Traction Control 2 Set ↑"},
		{"DiffPreloadLevel", "Differential Preload Set"},
		{"ShockLeftRearDec", "Left Rear Shock Set ↓"},
		{"MGUKRegenGainInc", "MGU-K Re-Gen Gain Set ↑"},
	}
	for _, tc := range cases {
		if got := Control(tc.id); got != tc.want {
			t.Errorf("Control(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestControlQuickChat(t *testing.T) {
	cases := []struct {
		id   string
		want string
	}{
		{"AutoChat1", "Pitting In"},
		{"AutoChat2", "Pitting Out"},
		{"AutoChat3", "Pass Left"},
		{"AutoChat4", "Pass Right"},
		{"AutoChat5", "Sorry!"},
		{"AutoChat6", "Thanks!"},
		{"AutoChat7", "AutoChat7"},
		{"AutoChat8", ":("},
		{"AutoChat9", "AutoChat9"},
		{"AutoChat12", "#rf rr"},
		{"AutoChat13", "#lf rf lr rr"},
		{"AutoChat14", "#fuel +0.5l"},
		{"AutoChat15", "#fuel -0.5l"},
	}
	for _, tc := range cases {
		if got := Control(tc.id); got != tc.want {
			t.Errorf("Control(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestControlAudioLabels(t *testing.T) {
	cases := []struct {
		id   string
		want string
	}{
		{"TrueForceVolumeUp", "Wheel LFE Louder"},
		{"TrueForceVolumeDown", "Wheel LFE Quieter"},
		{"LFEVolumeUp", "BassShaker LFE Louder"},
		{"LFEVolumeDown", "BassShaker LFE Quieter"},
	}
	for _, tc := range cases {
		if got := Control(tc.id); got != tc.want {
			t.Errorf("Control(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestKeyLabels(t *testing.T) {
	cases := []struct {
		code byte
		want string
	}{
		{'U', "U"},
		{'5', "5"},
		{0x20, "Space"},
		{0x2c, "Print Screen"},
		{0x2d, "Insert"},
		{0x2e, "Delete"},
		{0x13, "Pause"},
		{0x21, "Page Up"},
		{0x22, "Page Down"},
		{0x26, "Up"},
		{0x28, "Down"},
		{0x6b, ","},
		{0x6c, "-"},
		{0x6d, "."},
		{0x70, ";"},
		{0x72, "="},
		{0x76, "]"},
		{0x78, "["},
		{0x0c, "Numpad 5"},
		{0xb1, "Numpad 3"},
		{0xb2, "Numpad 1"},
		{0xb3, "Numpad 7"},
		{0xb4, "Numpad 4"},
		{0xb5, "Numpad 8"},
		{0xb6, "Numpad 6"},
		{0xb9, "Numpad ."},
		{0xc5, "F1"},
		{0xd0, "F12"},
		{0xfe, "0xFE"},
	}
	for _, tc := range cases {
		if got := Key(tc.code); got != tc.want {
			t.Errorf("Key(0x%02X) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestControlLabelsExistInFixtures(t *testing.T) {
	known := map[string]bool{}
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
			known[rec.Name] = true
		}
	}
	for id := range controlLabels {
		if !known[id] {
			t.Errorf("controlLabels contains %q, which is not in any cfgparse fixture", id)
		}
	}
}

func TestControlLabelsUnique(t *testing.T) {
	seen := map[string]string{}
	for id, label := range controlLabels {
		if prev, ok := seen[label]; ok {
			t.Errorf("duplicate label %q used by %q and %q", label, prev, id)
		}
		seen[label] = id
	}
}
