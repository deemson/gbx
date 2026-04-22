package git

import (
	"context"
	"strings"

	"github.com/deemson/gbx/internal/git/exec"
)

type Repo struct {
	path string
}

func (r Repo) Path() string {
	return r.path
}

func (r Repo) git() exec.Git {
	return exec.Git{Path: r.path}
}

func (r Repo) runGit(ctx context.Context, args ...string) (exec.Result, error) {
	return r.git().Run(ctx, args...)
}

func (r Repo) Branch(ctx context.Context) (string, error) {
	res, err := r.runGit(ctx, "branch", "--show-current")
	if err != nil {
		return "", NewUnknownRunErr(res, err)
	}
	return strings.TrimSpace(string(res.Stdout)), nil
}

func (r Repo) Status(ctx context.Context) (any, error) {
	res, err := r.runGit(ctx, "status", "--null", "--porcelain=v2")
	return nil, NewUnknownRunErr(res, err)
}
