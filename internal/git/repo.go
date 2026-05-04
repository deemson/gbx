package git

import (
	"bytes"
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

func (r Repo) RevParseHead(ctx context.Context) (string, error) {
	res, err := r.runGit(ctx, "rev-parse", "HEAD")
	if err != nil {
		return "", NewUnknownRunErr(res, err)
	}
	return string(bytes.TrimSpace(res.Stdout)), nil
}

func (r Repo) BranchShowCurrent(ctx context.Context) (string, error) {
	res, err := r.runGit(ctx, "branch", "--show-current")
	if err != nil {
		return "", NewUnknownRunErr(res, err)
	}
	return string(bytes.TrimSpace(res.Stdout)), nil
}

func (r Repo) Status(ctx context.Context) (Status, error) {
	res, err := r.runGit(ctx, "status", "-z", "--porcelain=v2", "--branch", "--show-stash")
	if err != nil {
		return Status{}, NewUnknownRunErr(res, err)
	}
	return parseStatus(res.Stdout)
}

func (r Repo) DiffNumStatHead(ctx context.Context) (DiffNumStat, error) {
	res, err := r.runGit(ctx, "diff", "HEAD", "-z", "--numstat")
	if err != nil {
		if res.ExitCode == 128 {
			stderr := string(res.Stderr)
			if strings.Contains(stderr, "'HEAD': unknown revision") {
				return DiffNumStat{}, ErrRepositoryHasNoCommits
			}
		}
		return DiffNumStat{}, NewUnknownRunErr(res, err)
	}
	return parseDiffNumStat(res.Stdout)
}
