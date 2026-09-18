package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type SwitchSuite struct {
	suite.Suite
}

func TestSwitchSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &SwitchSuite{})
}

func (s *SwitchSuite) TestOK() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	currentBranch := repo.BranchShowCurrent()
	repo.CheckoutBranch("branch")
	repo.Checkout(currentBranch)

	err := repo.Repo().Switch(ctx, "branch")
	if s.Assert().NoError(err) {
		s.Assert().Equal("branch", repo.BranchShowCurrent())
	}
}

func (s *SwitchSuite) TestNonExistent() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	err := repo.Repo().Switch(ctx, "non-existent")
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrUnknownPathspec)
		var runErr *git.RunError
		if s.Assert().ErrorAs(err, &runErr) {
			s.Assert().Contains(string(runErr.Res.Stderr), "invalid reference")
		}
	}
}

func (s *SwitchSuite) TestLocalChangesOverwritten() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	currentBranch := repo.BranchShowCurrent()
	repo.CheckoutBranch("branch")
	repo.WriteFileAdd("file", "branch data")
	repo.Commit("branch change")
	repo.Checkout(currentBranch)

	repo.WriteFile("file", "uncommitted")

	err := repo.Repo().Switch(ctx, "branch")
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrLocalChangesOverwritten)
	}
}

func (s *SwitchSuite) TestUntrackedOverwritten() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	currentBranch := repo.BranchShowCurrent()
	repo.CheckoutBranch("branch")
	repo.WriteFileAdd("untracked", "branch data")
	repo.Commit("add file")
	repo.Checkout(currentBranch)

	repo.WriteFile("untracked", "local data")

	err := repo.Repo().Switch(ctx, "branch")
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrUntrackedOverwritten)
	}
}
