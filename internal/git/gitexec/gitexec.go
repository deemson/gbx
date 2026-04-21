package gitexec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(ctx context.Context, path string, args ...string) (Result, error) {
	fullArgs := append([]string{"-C", path}, args...)
	cmd := exec.CommandContext(ctx, "git", fullArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	res := Result{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: cmd.ProcessState.ExitCode(),
	}
	if runErr != nil {
		return res, fmt.Errorf(
			"git %s: %w: stdout=`%s` stderr=`%s`",
			strings.Join(args, " "), runErr, res.Stdout, res.Stderr,
		)
	}
	return res, nil
}
