package config

import (
	"github.com/kaptinlin/gozod"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Actions Actions `toml:"actions" gozod:"required"`
}

type Actions struct {
	Enter      []string `toml:"enter" gozod:"required"`
	ShiftEnter []string `toml:"shift-enter" gozod:"required"`
}

func Load(data []byte) (Config, error) {
	var v any
	if err := toml.Unmarshal(data, &v); err != nil {
		return Config{}, err
	}
	schema := gozod.FromStruct[Config](gozod.WithFieldNameTag("toml"))
	jsonSchema, err := gozod.ToJSONSchema(schema)
	if err != nil {
		panic(err)
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
