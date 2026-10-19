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
