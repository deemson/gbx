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
	ErrUnknownPathspec        = errors.New("unknown pathspec")

	ErrLocalChangesOverwritten = errors.New("local changes would be overwritten by checkout")
	ErrUntrackedOverwritten    = errors.New("untracked files would be overwritten by checkout")

	ErrBranchAlreadyExists = errors.New("branch already exists")

	ErrNoRemote = errors.New("no remote")

	ErrNoUpstream     = errors.New("no tracking information for the current branch")
	ErrNotFastForward = errors.New("not possible to fast-forward")
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

type UnknownRunError struct {
	Res exec.Result
	Err error
}

func NewUnknownRunErr(res exec.Result, err error) *UnknownRunError {
	return &UnknownRunError{Res: res, Err: err}
}

func (e *UnknownRunError) Error() string {
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
