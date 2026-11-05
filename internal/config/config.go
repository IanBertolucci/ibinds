package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the resolved application configuration.
type Config struct {
	IRacingDir                string `json:"iRacingDir,omitempty"`
	ProfilesDir               string `json:"profilesDir,omitempty"`
	DefaultDeviceInstanceGUID string `json:"defaultDeviceInstanceGUID"`
	Onboarded                 bool   `json:"onboarded"`
}

// ProfilesDirFor returns the standard controls profiles directory for an
// iRacing folder: <iRacingDir>/profiles/controls.
func ProfilesDirFor(iRacingDir string) string {
	return filepath.Join(filepath.Clean(iRacingDir), "profiles", "controls")
}

// IRacingDirFromProfiles returns the iRacing folder when profilesDir follows
// the standard <root>/profiles/controls layout (case-insensitive on the last
// two components).
func IRacingDirFromProfiles(profilesDir string) (string, bool) {
	clean := filepath.Clean(profilesDir)
	if !strings.EqualFold(filepath.Base(clean), "controls") {
		return "", false
	}
	parent := filepath.Dir(clean)
	if !strings.EqualFold(filepath.Base(parent), "profiles") {
		return "", false
	}
	return filepath.Dir(parent), true
}

// FlagDir is the value of the --dir command-line flag. Empty means unset.
// Set by cmd/ibinds before calling Load.
var FlagDir string

// Load resolves the profiles directory with this priority:
// --dir flag > IBINDS_PROFILES_DIR env > user config file > default.
func Load() (*Config, error) {
	c, err := loadFile()
	if err != nil {
		c = &Config{}
	}
	if FlagDir != "" {
		c.ProfilesDir = filepath.Clean(FlagDir)
		if dir, ok := IRacingDirFromProfiles(c.ProfilesDir); ok {
			c.IRacingDir = dir
		}
		c.Onboarded = true
		return c, nil
	}
	if env := os.Getenv("IBINDS_PROFILES_DIR"); env != "" {
		c.ProfilesDir = filepath.Clean(env)
		if dir, ok := IRacingDirFromProfiles(c.ProfilesDir); ok {
			c.IRacingDir = dir
		}
		c.Onboarded = true
		return c, nil
	}
	if c.IRacingDir != "" {
		c.IRacingDir = filepath.Clean(c.IRacingDir)
		c.ProfilesDir = ProfilesDirFor(c.IRacingDir)
		c.Onboarded = true
		return c, nil
	}
	if c.ProfilesDir != "" {
		c.ProfilesDir = filepath.Clean(c.ProfilesDir)
		if dir, ok := IRacingDirFromProfiles(c.ProfilesDir); ok {
			c.IRacingDir = dir
		}
		c.Onboarded = true
		return c, nil
	}
	doc, err := DocumentsDir()
	if err != nil {
		return nil, err
	}
	c.IRacingDir = filepath.Join(doc, "iRacing")
	c.ProfilesDir = ProfilesDirFor(c.IRacingDir)
	return c, nil
}

// Save persists the config as JSON in the user config directory.
func Save(c *Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ibinds", "config.json"), nil
}

// LogPath returns the default log file location next to the config file.
func LogPath() (string, error) {
	p, err := configPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(p), "ibinds.log"), nil
}

func loadFile() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
