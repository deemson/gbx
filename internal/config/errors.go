package config

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var (
	ErrNotFound = errors.New("not found")
)

// FileExistsError reports that WriteDefault refused to overwrite an existing
// target file. Path is the offending file.
type FileExistsError struct {
	Path string
}

func (e *FileExistsError) Error() string {
	return e.Path + " already exists"
}

// TOMLError adapts a go-toml decode error into a concise one-line message with
// the position prefixed ("line R, column C: msg"), since DecodeError.Error()
// omits the position and DecodeError.String()'s multi-line code frame mangles
// multi-line spans. It unwraps to the underlying *toml.DecodeError.
type TOMLError struct {
	err *toml.DecodeError
}

func (e *TOMLError) Error() string {
	row, col := e.err.Position()
	msg := strings.TrimPrefix(e.err.Error(), "toml: ")
	return fmt.Sprintf("line %d, column %d: %s", row, col, msg)
}

func (e *TOMLError) Unwrap() error { return e.err }

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
	itemsBound  = regexp.MustCompile(`at (least|most) (\d+) items`)
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
		case "minItems", "maxItems":
			if m := itemsBound.FindStringSubmatch(msg); m != nil {
				issues = append(issues, Issue{toDotted(path), fmt.Sprintf("must have at %s %s items", m[1], m[2])})
			}
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
