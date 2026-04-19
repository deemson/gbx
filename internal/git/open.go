package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrDoesNotExist  = errors.New("path does not exist")
	ErrNotDirectory  = errors.New("not a directory")
	ErrNotRepository = errors.New("not a git repository")
)

func Open(ctx context.Context, path string) (Repo, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if cmd.ProcessState.ExitCode() == 128 {
			stderrString := strings.TrimSpace(stderr.String())
			if strings.Contains(stderrString, "not a git repository") {
				return Repo{}, ErrNotRepository
			}
			if strings.Contains(stderrString, "Not a directory") {
				return Repo{}, ErrNotDirectory
			}
			if strings.Contains(stderrString, "No such file or directory") {
				return Repo{}, ErrDoesNotExist
			}
		}
		return Repo{}, fmt.Errorf(
			"unknown error when opening repo %w: stdout=`%s` stderr=`%s`",
			err,
			strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String()),
		)
	}
	return Repo{}, nil
}
