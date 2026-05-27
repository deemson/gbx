package git

import (
	"context"
	"fmt"
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

func (r Repo) Checkout(ctx context.Context, what string) error {
	res, err := r.runGit(ctx, "checkout", what)
	if err != nil {
		if res.ExitCode == 1 {
			stderr := string(res.Stderr)
			switch {
			case strings.Contains(stderr, fmt.Sprintf("pathspec '%s' did not match", what)):
				return ErrUnknownPathspec
			case strings.Contains(stderr, "local changes to the following files would be overwritten"):
				return ErrLocalChangesOverwritten
			case strings.Contains(stderr, "untracked working tree files would be overwritten"):
				return ErrUntrackedOverwritten
			}
		}
		return NewUnknownRunErr(res, err)
	}
	return nil
}

func (r Repo) CheckoutBranch(ctx context.Context, name string) error {
	res, err := r.runGit(ctx, "checkout", "-b", name)
	if err != nil {
		if res.ExitCode == 128 {
			stderr := string(res.Stderr)
			if strings.Contains(stderr, fmt.Sprintf("a branch named '%s' already exists", name)) {
				return ErrBranchAlreadyExists
			}
		}
		return NewUnknownRunErr(res, err)
	}
	return nil
}

// Branches lists the repo's local branch names. A repo with no commits has no
// branches and yields an empty slice.
func (r Repo) Branches(ctx context.Context) ([]string, error) {
	res, err := r.runGit(ctx, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, NewUnknownRunErr(res, err)
	}
	var branches []string
	for line := range strings.SplitSeq(string(res.Stdout), "\n") {
		if b := strings.TrimSpace(line); b != "" {
			branches = append(branches, b)
		}
	}
	return branches, nil
}

func (r Repo) Fetch(ctx context.Context) error {
	res, err := r.runGit(ctx, "fetch")
	if err != nil {
		if res.ExitCode == 128 {
			stderr := string(res.Stderr)
			if strings.Contains(stderr, "Could not read from remote repository") {
				return ErrNoRemote
			}
		}
		return NewUnknownRunErr(res, err)
	}
	return nil
}

func (r Repo) Pull(ctx context.Context) error {
	res, err := r.runGit(ctx, "pull", "--ff-only")
	if err != nil {
		stderr := string(res.Stderr)
		switch {
		case res.ExitCode == 1 && strings.Contains(stderr, "There is no tracking information"):
			return ErrNoUpstream
		case res.ExitCode == 128 && strings.Contains(stderr, "Not possible to fast-forward"):
			return ErrNotFastForward
		case res.ExitCode == 1 && strings.Contains(stderr, "local changes to the following files would be overwritten"):
			return ErrLocalChangesOverwritten
		case res.ExitCode == 1 && strings.Contains(stderr, "untracked working tree files would be overwritten"):
			return ErrUntrackedOverwritten
		}
		return NewUnknownRunErr(res, err)
	}
	return nil
}
