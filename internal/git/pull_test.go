package git_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type PullSuite struct {
	suite.Suite
}

func TestPullSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &PullSuite{})
}

func (s *PullSuite) cloneWithRemote() (origin, clone gitest.Repo) {
	remoteDir := s.T().TempDir()
	gitest.InitBare(s.T(), remoteDir)

	origin = gitest.Init(s.T(), s.T().TempDir())
	origin.RemoteAdd("origin", remoteDir)
	origin.SetupCommitConfig()
	origin.WriteFileAdd("file", "v1")
	origin.Commit("c1")
	origin.PushSetUpstream("origin", origin.BranchShowCurrent())

	clone = gitest.Clone(s.T(), remoteDir, s.T().TempDir())
	clone.SetupCommitConfig()
	return origin, clone
}

func (s *PullSuite) TestOK() {
	ctx := context.Background()
	origin, clone := s.cloneWithRemote()

	origin.WriteFileAdd("file", "v2")
	origin.Commit("c2")
	origin.Push()

	err := clone.Repo().PullFastForward(ctx)
	s.Assert().NoError(err)
}

func (s *PullSuite) TestNoUpstream() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	err := repo.Repo().PullFastForward(ctx)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrNoUpstream)
	}
}

func (s *PullSuite) TestNotFastForward() {
	ctx := context.Background()
	origin, clone := s.cloneWithRemote()

	origin.WriteFileAdd("file", "v2")
	origin.Commit("c2")
	origin.Push()

	clone.WriteFileAdd("file", "diverged")
	clone.Commit("diverging commit")

	err := clone.Repo().PullFastForward(ctx)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrNotFastForward)
	}
}

func (s *PullSuite) TestMergeRefNotFetched() {
	ctx := context.Background()
	_, clone := s.cloneWithRemote()

	// Simulate the upstream branch having been deleted/renamed on the remote:
	// the branch's configured merge ref no longer exists there, so pull fetches
	// but never sees it.
	clone.Config(fmt.Sprintf("branch.%s.merge", clone.BranchShowCurrent()), "refs/heads/does-not-exist")

	err := clone.Repo().PullFastForward(ctx)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrMergeRefNotFetched)
	}
}

func (s *PullSuite) TestLocalChangesOverwritten() {
	ctx := context.Background()
	origin, clone := s.cloneWithRemote()

	origin.WriteFileAdd("file", "v2")
	origin.Commit("c2")
	origin.Push()

	clone.WriteFile("file", "uncommitted")

	err := clone.Repo().PullFastForward(ctx)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrLocalChangesOverwritten)
	}
}

func (s *PullSuite) TestUntrackedOverwritten() {
	ctx := context.Background()
	origin, clone := s.cloneWithRemote()

	origin.WriteFileAdd("untracked", "remote data")
	origin.Commit("add untracked")
	origin.Push()

	clone.WriteFile("untracked", "local data")

	err := clone.Repo().PullFastForward(ctx)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrUntrackedOverwritten)
	}
}
