package gitest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/exec"
	"github.com/stretchr/testify/require"
)

type Repo struct {
	repo git.Repo
	t    *testing.T
}

func (r Repo) Repo() git.Repo {
	return r.repo
}

func (r Repo) git() exec.Git {
	return exec.Git{Path: r.repo.Path()}
}

func (r Repo) runGit(args ...string) (exec.Result, error) {
	ctx := context.Background()
	return r.git().Run(ctx, args...)
}

func (r Repo) RevParseHead() string {
	commit, err := r.repo.RevParseHead(context.Background())
	require.NoError(r.t, err)
	return commit
}

func (r Repo) BranchShowCurrent() string {
	branch, err := r.repo.BranchShowCurrent(context.Background())
	require.NoError(r.t, err)
	return branch
}

func (r Repo) BranchSetUpstreamTo(remote, remoteBranch, branch string) {
	res, err := r.runGit("branch", fmt.Sprintf("--set-upstream-to=%s/%s", remote, remoteBranch), branch)
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) Checkout(what string) {
	res, err := r.runGit("checkout", what)
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) CheckoutBranch(name string) {
	res, err := r.runGit("checkout", "-b", name)
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) WriteFile(subPath string, data string) {
	require.NoError(r.t, os.WriteFile(path.Join(r.repo.Path(), subPath), []byte(data), 0644))
}

func (r Repo) RemovePath(subPath string) {
	require.NoError(r.t, os.Remove(path.Join(r.repo.Path(), subPath)))
}

func (r Repo) Add(subPath string) {
	res, err := r.runGit("add", subPath)
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) WriteFileAdd(subPath string, data string) {
	r.WriteFile(subPath, data)
	r.Add(subPath)
}

func (r Repo) RemovePathAdd(subPath string) {
	r.RemovePath(subPath)
	r.Add(subPath)
}

func (r Repo) SetupCommitConfig() {
	for _, args := range [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"config", "commit.gpgsign", "false"},
	} {
		if res, err := r.runGit(args...); err != nil {
			require.NoError(r.t, git.NewUnknownRunErr(res, err))
		}
	}
}

func (r Repo) Commit(message string) {
	res, err := r.runGit("commit", "-m", message)
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) Merge(what string) {
	res, err := r.runGit("merge", what)
	if err != nil {
		if res.ExitCode == 1 && bytes.Contains(res.Stdout, []byte("Automatic merge failed; fix conflicts")) {
			return
		}
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) Push() {
	res, err := r.runGit("push")
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) Pull() {
	res, err := r.runGit("pull")
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) Fetch() {
	res, err := r.runGit("fetch")
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) RemoteAdd(name, url string) {
	res, err := r.runGit("remote", "add", name, url)
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}

func (r Repo) PushSetUpstream(upstream, branch string) {
	res, err := r.runGit("push", "--set-upstream", upstream, branch)
	if err != nil {
		require.NoError(r.t, git.NewUnknownRunErr(res, err))
	}
}
