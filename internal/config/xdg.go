package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/deemson/gbx/internal/xdg"
	"github.com/pelletier/go-toml/v2"
)

const (
	xdgConfigRelPath = "gbx/config.toml"
	xdgSchemaRelPath = "gbx/schema.json"
)

func Find() (string, Config, error) {
	configPath, err := xdg.ConfigFile(xdgConfigRelPath)
	if err != nil {
		return "", Config{}, err
	}
	f, err := os.Open(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return configPath, Config{}, ErrNotFound
	}
	if err != nil {
		return configPath, Config{}, err
	}
	defer func() { _ = f.Close() }()
	cfg, err := FromReader(f)
	if err != nil {
		return configPath, Config{}, err
	}
	return configPath, cfg, nil
}

// WriteDefault writes the default config.toml and its companion schema.json to
// the primary XDG config dir, returning the paths written. Unless force is set,
// it refuses (writing nothing) and returns a *FileExistsError if either target
// already exists.
func WriteDefault(force bool) ([]string, error) {
	configPath, err := xdg.ConfigFile(xdgConfigRelPath)
	if err != nil {
		return nil, err
	}
	schemaPath, err := xdg.ConfigFile(xdgSchemaRelPath)
	if err != nil {
		return nil, err
	}
	if !force {
		for _, path := range []string{configPath, schemaPath} {
			if _, err := os.Stat(path); err == nil {
				return nil, &FileExistsError{Path: path}
			}
		}
	}
	jsonSchemaData, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		return nil, err
	}
	configData, err := toml.Marshal(Default())
	if err != nil {
		return nil, err
	}
	// Prepend the schema directive editors (Taplo, etc.) read to validate against
	// the companion schema.json written alongside this file.
	configData = append([]byte("#:schema ./"+filepath.Base(xdgSchemaRelPath)+"\n"), configData...)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(schemaPath, jsonSchemaData, 0644); err != nil {
		return nil, err
	}
	return []string{configPath, schemaPath}, nil
}
