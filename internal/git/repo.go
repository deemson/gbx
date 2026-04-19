package git

import (
	"context"
	"os/exec"
)

type Repo struct {
	path string
}

func (r Repo) cmd(ctx context.Context, args ...string) *exec.Cmd {
	baseArgs := []string{"-C", r.path}
	return exec.CommandContext(ctx, "git", append(baseArgs, args...)...)
}

func (r Repo) Status(ctx context.Context) {
	cmd := r.cmd(ctx, "status", "--porcelain")
	_ = cmd
}

type TestRepo struct {
	Repo
}


