package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is persisted at ~/.huginn/config.json
type Config struct {
	Version int    `json:"version"`
	Theme   string `json:"theme"`
	Memory  struct {
		Enabled    bool `json:"enabled"`
		Learning   bool `json:"learning"`
		AutoSave   bool `json:"auto_save"`
		MaxResults int  `json:"max_results"`
	} `json:"memory"`
	Agents struct {
		Default string `json:"default"`
	} `json:"agents"`
}

func Default() Config {
	var c Config
	c.Version = 1
	c.Theme = "vikingpunk"
	c.Memory.Enabled = true
	c.Memory.Learning = true
	c.Memory.AutoSave = false
	c.Memory.MaxResults = 10
	c.Agents.Default = "planner"
	return c
}

func Path(baseDir string) string {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".huginn")
	}
	return filepath.Join(baseDir, "config.json")
}

func Load(baseDir string) (Config, error) {
	p := Path(baseDir)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Default(), err
	}
	if c.Version == 0 {
		c = Default()
	}
	return c, nil
}

func Save(baseDir string, c Config) error {
	if err := os.MkdirAll(filepath.Dir(Path(baseDir)), 0755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(Path(baseDir), b, 0644)
}
