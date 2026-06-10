package config

import (
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
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
	err := toml.Unmarshal(data, &v)
	if err != nil {
		return Config{}, err
	}
	schema := gozod.FromStruct[Config](gozod.WithFieldNameTag("toml"))
	jsonSchema, err := gozod.ToJSONSchema(schema)
	if err != nil {
		panic(err)
	}
	jsData, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsData))
	ev := jsonSchema.Validate(v)
	spew.Dump(ev.ToList())
	return Config{}, nil
}
