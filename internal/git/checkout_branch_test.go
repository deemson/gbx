package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type CheckoutBranchSuite struct {
	suite.Suite
}

func TestCheckoutBranchSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CheckoutBranchSuite{})
}

func (s *CheckoutBranchSuite) TestOK() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	err := repo.Repo().CheckoutBranch(ctx, "branch")
	if s.Assert().NoError(err) {
		s.Assert().Equal("branch", repo.BranchShowCurrent())
	}
}

func (s *CheckoutBranchSuite) TestAlreadyExists() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	err := repo.Repo().CheckoutBranch(ctx, "branch")
	if s.Assert().NoError(err) {
		err = repo.Repo().CheckoutBranch(ctx, "branch")
		if s.Assert().Error(err) {
			s.Assert().ErrorIs(err, git.ErrBranchAlreadyExists)
		}
	}
}
