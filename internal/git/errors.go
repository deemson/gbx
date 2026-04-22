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
