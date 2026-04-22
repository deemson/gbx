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
)

type UnknownError struct {
	Res exec.Result
	Err error
}

func NewErrUnknown(res exec.Result, err error) *UnknownError {
	return &UnknownError{Res: res, Err: err}
}

func (e *UnknownError) Error() string {
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
