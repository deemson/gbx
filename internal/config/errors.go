package config

import (
	"errors"
	"fmt"
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
