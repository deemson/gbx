package config

import (
	"errors"
	"io"

	"github.com/pelletier/go-toml/v2"
)

func FromBytes(data []byte) (Config, error) {
	var v any
	if err := toml.Unmarshal(data, &v); err != nil {
		var decErr *toml.DecodeError
		if errors.As(err, &decErr) {
			return Config{}, &TOMLError{decErr}
		}
		return Config{}, err
	}
	if err := validate(v); err != nil {
		return Config{}, err
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
