package config

import (
	"encoding/json"
	"os"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

const (
	xdgConfigRelPath = "gbx/config.toml"
	xdgSchemaRelPath = "gbx/schema.json"
)

func Find() (string, Config, error) {
	configPath, err := xdg.SearchConfigFile(xdgConfigRelPath)
	if err != nil {
		return "", Config{}, ErrNotFound
	}
	f, err := os.Open(configPath)
	if err != nil {
		return configPath, Config{}, err
	}
	defer f.Close()
	cfg, err := FromReader(f)
	if err != nil {
		return configPath, Config{}, err
	}
	return configPath, cfg, nil
}

func WriteDefault() error {
	jsonSchemaData, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		return err
	}
	configData, err := toml.Marshal(Default())
	if err != nil {
		return err
	}
	schemaPath, err := xdg.ConfigFile(xdgSchemaRelPath)
	if err != nil {
		return err
	}
	configPath, err := xdg.ConfigFile(xdgConfigRelPath)
	if err != nil {
		return err
	}
	err = os.WriteFile(schemaPath, jsonSchemaData, 0644)
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		return err
	}
	return nil
}
