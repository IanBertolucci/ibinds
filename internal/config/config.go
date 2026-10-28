package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the resolved application configuration.
type Config struct {
	ProfilesDir               string `json:"profilesDir"`
	DefaultDeviceInstanceGUID string `json:"defaultDeviceInstanceGUID"`
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
		c.ProfilesDir = FlagDir
		return c, nil
	}
	if env := os.Getenv("IBINDS_PROFILES_DIR"); env != "" {
		c.ProfilesDir = env
		return c, nil
	}
	if c.ProfilesDir != "" {
		return c, nil
	}
	doc, err := DocumentsDir()
	if err != nil {
		return nil, err
	}
	c.ProfilesDir = filepath.Join(doc, "iRacing", "profiles", "controls")
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
