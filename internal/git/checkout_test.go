package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type CheckoutSuite struct {
	suite.Suite
}

func TestCheckoutSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CheckoutSuite{})
}

func (s *CheckoutSuite) TestOK() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	currentBranch := repo.BranchShowCurrent()
	anotherBranch := "branch"
	repo.CheckoutBranch(anotherBranch)

	err := repo.Repo().Checkout(ctx, currentBranch)
	if s.Assert().NoError(err) {
		err = repo.Repo().Checkout(ctx, anotherBranch)
		s.Assert().NoError(err)
	}
}

func (s *CheckoutSuite) TestNonExistent() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	err := repo.Repo().Checkout(ctx, "non-existent")
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrUnknownPathspec)
	}
}

func (s *CheckoutSuite) TestLocalChangesOverwritten() {
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

	err := repo.Repo().Checkout(ctx, "branch")
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrLocalChangesOverwritten)
	}
}

func (s *CheckoutSuite) TestUntrackedOverwritten() {
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

	err := repo.Repo().Checkout(ctx, "branch")
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrUntrackedOverwritten)
	}
}
