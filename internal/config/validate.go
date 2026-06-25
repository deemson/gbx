package config

import (
	"fmt"
	"sort"
)

// validate walks the generically-decoded TOML tree and collects every problem
// at once as a sorted ValidationError. It checks the full structure deeply
// enough that, once it passes, unmarshaling into Config cannot hit a type
// mismatch. It mirrors the rules in jsonSchema (schema.go).
func validate(v any) error {
	var issues []Issue
	root, ok := v.(map[string]any)
	if !ok {
		// TOML always decodes to a table at the root; guard regardless.
		issues = append(issues, Issue{"", fmt.Sprintf("expected object, got %s", typeName(v))})
		return asError(issues)
	}
	for key := range root {
		if key != "actions" {
			issues = append(issues, Issue{key, "unknown field"})
		}
	}
	actions, ok := root["actions"]
	if !ok {
		issues = append(issues, Issue{"actions", "required field is missing"})
		return asError(issues)
	}
	issues = append(issues, validateActions(actions)...)
	return asError(issues)
}

func validateActions(v any) []Issue {
	arr, ok := v.([]any)
	if !ok {
		return []Issue{{"actions", fmt.Sprintf("expected array, got %s", typeName(v))}}
	}
	switch {
	case len(arr) < 1:
		return []Issue{{"actions", "must have at least 1 items"}}
	case len(arr) > 9:
		return []Issue{{"actions", "must have at most 9 items"}}
	}
	var issues []Issue
	for i, item := range arr {
		issues = append(issues, validateAction(fmt.Sprintf("actions.%d", i), item)...)
	}
	return issues
}

func validateAction(path string, v any) []Issue {
	m, ok := v.(map[string]any)
	if !ok {
		return []Issue{{path, fmt.Sprintf("expected object, got %s", typeName(v))}}
	}
	var issues []Issue
	for key := range m {
		if key != "label" && key != "command" {
			issues = append(issues, Issue{path + "." + key, "unknown field"})
		}
	}
	if label, ok := m["label"]; !ok {
		issues = append(issues, Issue{path + ".label", "required field is missing"})
	} else if _, ok := label.(string); !ok {
		issues = append(issues, Issue{path + ".label", fmt.Sprintf("expected string, got %s", typeName(label))})
	}
	if command, ok := m["command"]; !ok {
		issues = append(issues, Issue{path + ".command", "required field is missing"})
	} else {
		issues = append(issues, validateCommand(path+".command", command)...)
	}
	return issues
}

func validateCommand(path string, v any) []Issue {
	arr, ok := v.([]any)
	if !ok {
		return []Issue{{path, fmt.Sprintf("expected array, got %s", typeName(v))}}
	}
	if len(arr) < 1 {
		return []Issue{{path, "must have at least 1 items"}}
	}
	var issues []Issue
	for i, item := range arr {
		if _, ok := item.(string); !ok {
			issues = append(issues, Issue{fmt.Sprintf("%s.%d", path, i), fmt.Sprintf("expected string, got %s", typeName(item))})
		}
	}
	return issues
}

func asError(issues []Issue) error {
	if len(issues) == 0 {
		return nil
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].String() < issues[j].String() })
	return &ValidationError{Issues: issues}
}

// typeName maps a go-toml generically-decoded value to its JSON-schema type
// name, for "expected X, got Y" messages.
func typeName(v any) string {
	switch v.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case bool:
		return "boolean"
	case int64:
		return "integer"
	case float64:
		return "number"
	default:
		return "unknown"
	}
}
