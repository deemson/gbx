package config

import (
	"io"

	"github.com/pelletier/go-toml/v2"
)

func FromBytes(data []byte) (Config, error) {
	var v any
	if err := toml.Unmarshal(data, &v); err != nil {
		return Config{}, err
	}
	if ev := jsonSchema.Validate(v); !ev.IsValid() {
		return Config{}, newValidationError(ev.DetailedErrors())
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func FromReader(r io.Reader) (Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Config{}, err
	}
	return FromBytes(data)
}
