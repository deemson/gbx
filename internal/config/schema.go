package config

import (
	"github.com/kaptinlin/gozod"
	"github.com/kaptinlin/jsonschema"
)

var jsonSchema = func() *jsonschema.Schema {
	schema := gozod.FromStruct[Config](gozod.WithFieldNameTag("toml"))
	jsonSchema, err := gozod.ToJSONSchema(schema)
	if err != nil {
		panic(err)
	}
	return jsonSchema
}()
