package config

// jsonSchema is the static JSON Schema for the config, marshaled to the
// companion schema.json by WriteDefault so editors (Taplo, etc.) validate the
// TOML. It mirrors the rules enforced at runtime by validate; the two are
// maintained by hand and the config scope is deliberately small and fixed.
var jsonSchema = map[string]any{
	"$schema":              "https://json-schema.org/draft/2020-12/schema",
	"type":                 "object",
	"additionalProperties": false,
	"required":             []any{"actions"},
	"properties": map[string]any{
		"actions": map[string]any{
			"type":     "array",
			"minItems": 1,
			"maxItems": 9,
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []any{"label", "command"},
				"properties": map[string]any{
					"label": map[string]any{"type": "string"},
					"command": map[string]any{
						"type":     "array",
						"minItems": 1,
						"items":    map[string]any{"type": "string"},
					},
				},
			},
		},
	},
}
