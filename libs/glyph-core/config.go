package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all glyph configuration values.
type Config struct {
	AIProvider   string `toml:"ai_provider"`
	AIModel      string `toml:"ai_model"`
	APIKey       string `toml:"api_key"`
	OllamaHost   string `toml:"ollama_host"`
	DefaultStyle string `toml:"default_style"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		AIProvider:   "groq",
		AIModel:      "llama-3.3-70b-versatile",
		OllamaHost:   "http://localhost:11434",
		DefaultStyle: "rounded",
	}
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate config dir: %w", err)
	}
	return filepath.Join(cfgDir, "glyph", "config.toml"), nil
}

// LoadConfig reads config from ~/.config/glyph/config.toml.
// If the file does not exist, defaults are returned without error.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	path, err := configPath()
	if err != nil {
		return cfg, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, &AppError{
			Msg: "failed to parse config file",
			Err: err,
		}
	}
	return cfg, nil
}

// WriteConfig persists cfg to ~/.config/glyph/config.toml,
// creating the directory if needed.
func WriteConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &AppError{
			Msg: "cannot create config directory",
			Err: err,
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return &AppError{
			Msg: "cannot write config file",
			Err: err,
		}
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		return &AppError{
			Msg: "cannot encode config",
			Err: err,
		}
	}
	return nil
}
