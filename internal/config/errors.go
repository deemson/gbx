package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Issue is a single human-readable config problem at a dotted path.
type Issue struct {
	Path    string
	Message string
}

func (i Issue) String() string {
	return i.Path + ": " + i.Message
}

// ValidationError reports config values that don't satisfy the schema, as a
// sorted list of plain-language issues (one per line).
type ValidationError struct {
	Issues []Issue
}

func (e *ValidationError) Error() string {
	lines := make([]string, len(e.Issues))
	for i, issue := range e.Issues {
		lines[i] = issue.String()
	}
	return strings.Join(lines, "\n")
}

var (
	quotedNames = regexp.MustCompile(`'([^']+)'`)
	valueType   = regexp.MustCompile(`Value is (\S+) but should be (\S+)`)
)

// newValidationError translates the JSON-schema evaluation's flat
// "instancePath/keyword -> message" map into plain-language issues. The schema
// uses additionalProperties:false, so unknown fields surface as a "false"
// subschema match (keyword "schema") whose path is the offending field; the
// "additionalProperties"/"properties" summary keywords are dropped as
// redundant, and a "null type" error is dropped in favour of its "required"
// twin.
func newValidationError(detailed map[string]string) *ValidationError {
	var issues []Issue
	for key, msg := range detailed {
		path, keyword := splitKey(key)
		switch keyword {
		case "schema":
			issues = append(issues, Issue{toDotted(path), "unknown field"})
		case "required":
			for _, m := range quotedNames.FindAllStringSubmatch(msg, -1) {
				issues = append(issues, Issue{joinPath(path, m[1]), "required field is missing"})
			}
		case "type":
			m := valueType.FindStringSubmatch(msg)
			if m == nil || m[1] == "null" {
				continue // null type is redundant with the "required" issue
			}
			issues = append(issues, Issue{toDotted(path), fmt.Sprintf("expected %s, got %s", m[2], m[1])})
		}
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].String() < issues[j].String() })
	return &ValidationError{Issues: issues}
}

// splitKey splits a detailed-errors key into its instance path and trailing
// keyword, e.g. "/actions/enter/type" -> ("/actions/enter", "type") and
// "required" -> ("", "required").
func splitKey(key string) (path, keyword string) {
	if i := strings.LastIndex(key, "/"); i >= 0 {
		return key[:i], key[i+1:]
	}
	return "", key
}

// toDotted turns an instance location ("/actions/enter") into a config path
// ("actions.enter").
func toDotted(path string) string {
	return strings.ReplaceAll(strings.TrimPrefix(path, "/"), "/", ".")
}

func joinPath(path, name string) string {
	if parent := toDotted(path); parent != "" {
		return parent + "." + name
	}
	return name
}
