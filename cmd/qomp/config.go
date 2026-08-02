package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "qomp.json"

// Config is project state written by qomp init / add / update.
type Config struct {
	Registry   string            `json:"registry"`
	Theme      string            `json:"theme"`
	Accent     string            `json:"accent,omitempty"`
	Paths      ConfigPaths       `json:"paths"`
	Components []string          `json:"components"`
	Hashes     map[string]string `json:"hashes,omitempty"`
}

type ConfigPaths struct {
	Components string `json:"components"`
	Styles     string `json:"styles"`
}

func defaultConfig(registry string) Config {
	return Config{
		Registry: registry,
		Theme:    "claude",
		Paths: ConfigPaths{
			Components: "components/qompnt",
			Styles:     "components/qompnt/styles",
		},
		Components: []string{},
	}
}

// findConfig walks from dir upward looking for qomp.json.
func findConfig(dir string) (cfg Config, root string, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return Config{}, "", err
	}
	for {
		path := filepath.Join(dir, configFileName)
		data, readErr := os.ReadFile(path)
		if readErr == nil {
			if err := json.Unmarshal(data, &cfg); err != nil {
				return Config{}, "", fmt.Errorf("%s: %w", path, err)
			}
			return cfg, dir, nil
		}
		if !errors.Is(readErr, os.ErrNotExist) {
			return Config{}, "", readErr
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Config{}, "", fmt.Errorf("no %s found (run qomp init first)", configFileName)
		}
		dir = parent
	}
}

func saveConfig(root string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(root, configFileName), data, 0o644)
}

func configExists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, configFileName))
	return err == nil
}
