package git

import (
	"errors"
	"fmt"
	"strings"

	"github.com/deemson/gbx/internal/git/exec"
)

var (
	ErrDoesNotExist  = errors.New("path does not exist")
	ErrNotDirectory  = errors.New("not a directory")
	ErrNotRepository = errors.New("not a git repository")

	ErrDotGitOpen = errors.New("attempt to open .git/ as a repository")

	ErrRepositoryHasNoCommits = errors.New("repository has no commits")
	ErrUnknownPathspec        = errors.New("no such branch")

	ErrLocalChangesOverwritten = errors.New("commit or stash changes first")
	ErrUntrackedOverwritten    = errors.New("untracked files in the way")

	ErrBranchAlreadyExists = errors.New("branch already exists")

	ErrNoRemote = errors.New("no remote configured")

	ErrNoUpstream         = errors.New("branch tracks no remote")
	ErrNotFastForward     = errors.New("diverged, can't fast-forward")
	ErrMergeRefNotFetched = errors.New("upstream missing, fetch first")
)

type TokenParseError struct {
	TokenIndex int
	Token      []byte
	Err        error
}

func (e TokenParseError) Error() string {
	errString := "<nil>"
	if e.Err != nil {
		errString = e.Err.Error()
	}
	return fmt.Sprintf("%s: token %d `%s`", errString, e.TokenIndex, string(e.Token))
}

type ParseError struct {
	Errs []error
}

func (e *ParseError) Error() string {
	errStrings := make([]string, len(e.Errs))
	for i, err := range e.Errs {
		errString := "<nil>"
		if err != nil {
			errString = err.Error()
		}
		errStrings[i] = errString
	}
	return strings.Join(errStrings, "; ")
}

// RunError preserves the complete result of a failed git invocation. Cause is
// the friendly sentinel used for a recognized failure; when it is nil, Err is
// the underlying process error and Error retains the legacy diagnostic string.
type RunError struct {
	Res   exec.Result
	Err   error
	Cause error
}

func NewRunErr(res exec.Result, err, cause error) *RunError {
	return &RunError{Res: res, Err: err, Cause: cause}
}

// UnknownRunError remains an alias for callers that used the old name. Both
// recognized and unrecognized failures now travel through RunError.
type UnknownRunError = RunError

func NewUnknownRunErr(res exec.Result, err error) *RunError {
	return NewRunErr(res, err, nil)
}

func (e *RunError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	errString := "<nil>"
	if e.Err != nil {
		errString = e.Err.Error()
	}
	return fmt.Sprintf(
		"%s: %s: stdout=`%s` stderr=`%s`",
		strings.Join(e.Res.Args, " "),
		errString,
		strings.TrimSpace(string(e.Res.Stdout)),
		strings.TrimSpace(string(e.Res.Stderr)),
	)
}

// Unwrap keeps errors.Is working for recognized sentinels and exposes the
// process failure for an unrecognized invocation.
func (e *RunError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return e.Err
}

// Summary is the concise explanation intended for a row or a details field.
func (e *RunError) Summary() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "<nil>"
}
