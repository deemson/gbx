package exec

import (
	"bytes"
	"context"
	"os/exec"
)

type Git struct {
	Path string
}

func (g Git) cmd(ctx context.Context, args ...string) ([]string, *exec.Cmd) {
	fullArgs := args
	if g.Path != "" {
		fullArgs = append([]string{"-C", g.Path}, args...)
	}
	return fullArgs, exec.CommandContext(ctx, "git", fullArgs...)
}

func (g Git) Run(ctx context.Context, args ...string) (Result, error) {
	fullArgs, cmd := g.cmd(ctx, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := Result{
		Args:     fullArgs,
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: cmd.ProcessState.ExitCode(),
	}
	return res, err
}
