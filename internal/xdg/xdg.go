// Package xdg resolves config/state file paths using the freedesktop base-dir
// layout on every platform (unlike github.com/adrg/xdg, which follows Apple's
// ~/Library/Application Support convention on macOS). It honors the $XDG_*_HOME
// env vars when set non-empty and falls back to ~/.config and ~/.local/state.
// The functions only compute paths; callers that write create the parent dir.
package xdg

import (
	"os"
	"path/filepath"
)

// ConfigFile resolves rel under $XDG_CONFIG_HOME (or ~/.config when unset/empty).
func ConfigFile(rel string) (string, error) {
	return resolve("XDG_CONFIG_HOME", ".config", rel)
}

// StateFile resolves rel under $XDG_STATE_HOME (or ~/.local/state when unset/empty).
func StateFile(rel string) (string, error) {
	return resolve("XDG_STATE_HOME", filepath.Join(".local", "state"), rel)
}

func resolve(envVar, homeRel, rel string) (string, error) {
	base := os.Getenv(envVar)
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, homeRel)
	}
	return filepath.Join(base, rel), nil
}
