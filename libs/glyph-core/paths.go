package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths provides XDG-compliant directory helpers for a named tool.
type Paths struct {
	toolName string
}

// NewPaths returns a Paths helper for toolName (e.g. "pin", "ask").
func NewPaths(toolName string) Paths {
	return Paths{toolName: toolName}
}

// DataDir returns ~/.local/share/glyph/<toolname> and ensures it exists.
func (p Paths) DataDir() (string, error) {
	base, err := xdgDataHome()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "glyph", p.toolName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", &AppError{
			Msg: fmt.Sprintf("cannot create data directory for %s", p.toolName),
			Err: err,
		}
	}
	return dir, nil
}

// ConfigDir returns ~/.config/glyph and ensures it exists.
func (p Paths) ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", &AppError{Msg: "cannot locate config dir", Err: err}
	}
	dir := filepath.Join(base, "glyph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", &AppError{Msg: "cannot create config directory", Err: err}
	}
	return dir, nil
}

// xdgDataHome returns $XDG_DATA_HOME or ~/.local/share.
func xdgDataHome() (string, error) {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", &AppError{Msg: "cannot locate home directory", Err: err}
	}
	return filepath.Join(home, ".local", "share"), nil
}
