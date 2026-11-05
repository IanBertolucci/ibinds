package config

import (
	"os"
	"path/filepath"
	"testing"
)

func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("IBINDS_PROFILES_DIR", "")
	FlagDir = ""
	return home
}

func TestDefaultProfilesDir(t *testing.T) {
	home := isolateHome(t)

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := filepath.Join(home, "Documents", "iRacing", "profiles", "controls")
	if c.ProfilesDir != want {
		t.Fatalf("ProfilesDir = %q, want %q", c.ProfilesDir, want)
	}

	if err := Save(c); err != nil {
		t.Fatalf("Save: %v", err)
	}
	p := filepath.Join(home, ".config", "ibinds", "config.json")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("config.json not written to %q: %v", p, err)
	}
}

func TestEnvOverride(t *testing.T) {
	home := isolateHome(t)
	dir := filepath.Join(home, "custom")
	t.Setenv("IBINDS_PROFILES_DIR", dir)

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.ProfilesDir != dir {
		t.Fatalf("ProfilesDir = %q, want %q", c.ProfilesDir, dir)
	}
}

func TestFileOverride(t *testing.T) {
	home := isolateHome(t)
	dir := filepath.Join(home, "fromfile")
	if err := Save(&Config{ProfilesDir: dir}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.ProfilesDir != dir {
		t.Fatalf("ProfilesDir = %q, want %q", c.ProfilesDir, dir)
	}
}

func TestDefaultDevicePersists(t *testing.T) {
	isolateHome(t)
	dev := "{991805D0-2EB6-11F1-800A-444553540000}"
	if err := Save(&Config{
		ProfilesDir:               filepath.Join(t.TempDir(), "profiles"),
		DefaultDeviceInstanceGUID: dev,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.DefaultDeviceInstanceGUID != dev {
		t.Fatalf("DefaultDeviceInstanceGUID = %q, want %q", c.DefaultDeviceInstanceGUID, dev)
	}
}

func TestDefaultDeviceSurvivesEnvOverride(t *testing.T) {
	home := isolateHome(t)
	dev := "{991805D0-2EB6-11F1-800A-444553540000}"
	if err := Save(&Config{
		ProfilesDir:               filepath.Join(home, "fromfile"),
		DefaultDeviceInstanceGUID: dev,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	envDir := filepath.Join(home, "env")
	t.Setenv("IBINDS_PROFILES_DIR", envDir)

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.ProfilesDir != envDir {
		t.Fatalf("ProfilesDir = %q, want %q", c.ProfilesDir, envDir)
	}
	if c.DefaultDeviceInstanceGUID != dev {
		t.Fatalf("DefaultDeviceInstanceGUID = %q, want %q", c.DefaultDeviceInstanceGUID, dev)
	}
}

func TestFlagOverride(t *testing.T) {
	home := isolateHome(t)
	t.Setenv("IBINDS_PROFILES_DIR", filepath.Join(home, "env"))
	if err := Save(&Config{ProfilesDir: filepath.Join(home, "file")}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	flagDir := filepath.Join(home, "flag")
	FlagDir = flagDir

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.ProfilesDir != flagDir {
		t.Fatalf("ProfilesDir = %q, want %q", c.ProfilesDir, flagDir)
	}
}

func TestDefaultIsFirstRun(t *testing.T) {
	home := isolateHome(t)
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	wantIRacing := filepath.Join(home, "Documents", "iRacing")
	if c.IRacingDir != wantIRacing {
		t.Fatalf("IRacingDir = %q, want %q", c.IRacingDir, wantIRacing)
	}
	if c.ProfilesDir != ProfilesDirFor(wantIRacing) {
		t.Fatalf("ProfilesDir = %q, want derived", c.ProfilesDir)
	}
	if c.Onboarded {
		t.Fatal("Onboarded = true, want false for a fresh config")
	}
}

func TestIRacingDirDerivesProfilesDir(t *testing.T) {
	home := isolateHome(t)
	root := filepath.Join(home, "Documents", "iRacing")
	if err := Save(&Config{IRacingDir: root, Onboarded: true}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.ProfilesDir != filepath.Join(root, "profiles", "controls") {
		t.Fatalf("ProfilesDir = %q, want derived", c.ProfilesDir)
	}
	if !c.Onboarded {
		t.Fatal("Onboarded = false, want true")
	}
}

func TestLegacyProfilesDirMigratesWithoutOnboardingModal(t *testing.T) {
	home := isolateHome(t)
	root := filepath.Join(home, "Documents", "iRacing")
	legacy := filepath.Join(root, "profiles", "controls")
	if err := Save(&Config{ProfilesDir: legacy}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.ProfilesDir != legacy {
		t.Fatalf("ProfilesDir = %q, want %q", c.ProfilesDir, legacy)
	}
	if c.IRacingDir != root {
		t.Fatalf("IRacingDir = %q, want %q", c.IRacingDir, root)
	}
	if !c.Onboarded {
		t.Fatal("Onboarded = false, want true for a legacy config")
	}
}

func TestLegacyCustomProfilesDirKeepsNoIRacingDir(t *testing.T) {
	home := isolateHome(t)
	custom := filepath.Join(home, "custom-controls")
	if err := Save(&Config{ProfilesDir: custom}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.ProfilesDir != custom {
		t.Fatalf("ProfilesDir = %q, want %q", c.ProfilesDir, custom)
	}
	if c.IRacingDir != "" {
		t.Fatalf("IRacingDir = %q, want empty", c.IRacingDir)
	}
	if !c.Onboarded {
		t.Fatal("Onboarded = false, want true")
	}
}

func TestOverrideClearsStaleIRacingDir(t *testing.T) {
	home := isolateHome(t)
	if err := Save(&Config{
		IRacingDir: filepath.Join(home, "saved-iracing"),
		Onboarded:  true,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	override := filepath.Join(home, "custom-controls")
	FlagDir = override

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.ProfilesDir != override {
		t.Fatalf("ProfilesDir = %q, want override %q", c.ProfilesDir, override)
	}
	if c.IRacingDir != "" {
		t.Fatalf("IRacingDir = %q, want empty for a non-derivable override", c.IRacingDir)
	}
}

func TestFlagOverrideDerivesIRacingDirAndOnboards(t *testing.T) {
	home := isolateHome(t)
	root := filepath.Join(home, "Documents", "iRacing")
	FlagDir = filepath.Join(root, "profiles", "controls")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.IRacingDir != root {
		t.Fatalf("IRacingDir = %q, want %q", c.IRacingDir, root)
	}
	if !c.Onboarded {
		t.Fatal("Onboarded = false, want true for flag override")
	}
}
