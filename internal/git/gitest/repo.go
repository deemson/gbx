package gitest

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/exec"
)

type Repo struct {
	git.Repo
}

func (r Repo) git() exec.Git {
	return exec.Git{Path: r.Path()}
}

func (r Repo) runGit(ctx context.Context, args ...string) (exec.Result, error) {
	return r.git().Run(ctx, args...)
}

func (r Repo) Checkout(ctx context.Context, what string) error {
	res, err := r.git().Run(ctx, "checkout", what)
	if err != nil {
		return git.NewUnknownRunErr(res, err)
	}
	return nil
}

func (r Repo) CheckoutBranch(ctx context.Context, name string) error {
	res, err := r.git().Run(ctx, "checkout", "-b", name)
	if err != nil {
		return git.NewUnknownRunErr(res, err)
	}
	return nil
}

func (r Repo) WriteFile(subPath string, data []byte) error {
	return os.WriteFile(path.Join(r.Path(), subPath), data, 0644)
}

func (r Repo) RemovePath(subPath string) error {
	return os.Remove(path.Join(r.Path(), subPath))
}

func (r Repo) Add(ctx context.Context, subPath string) error {
	res, err := r.runGit(ctx, "add", subPath)
	if err != nil {
		return git.NewUnknownRunErr(res, err)
	}
	return nil
}

func (r Repo) SetupCommitConfig(ctx context.Context) error {
	g := r.git()
	for _, args := range [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"config", "commit.gpgsign", "false"},
	} {
		if res, err := g.Run(ctx, args...); err != nil {
			return git.NewUnknownRunErr(res, err)
		}
	}
	return nil
}

func (r Repo) Commit(ctx context.Context, message string) error {
	res, err := r.git().Run(ctx, "commit", "-m", message)
	if err != nil {
		return git.NewUnknownRunErr(res, err)
	}
	return nil
}

func (r Repo) Merge(ctx context.Context, what string) error {
	res, err := r.git().Run(ctx, "merge", what)
	if err != nil {
		if res.ExitCode == 1 && strings.Contains(string(res.Stdout), "Automatic merge failed; fix conflicts") {
			return nil
		}
		return git.NewUnknownRunErr(res, err)
	}
	return nil
}

func (r Repo) Git(ctx context.Context, args ...string) (exec.Result, error) {
	return exec.Git{Path: r.Path()}.Run(ctx, args...)
}
