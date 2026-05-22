package git_test

import (
	"context"
	"testing"

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

func (s *SwitchSuite) TestSwitchToExistingBranch() {
	ctx := context.Background()

	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "v1\n")
	repo.Commit("c1")
	start := repo.BranchShowCurrent()
	repo.CheckoutBranch("feature")
	repo.Checkout(start)

	s.Require().NoError(repo.Repo().Switch(ctx, "feature"))
	s.Assert().Equal("feature", repo.BranchShowCurrent())
}

func (s *SwitchSuite) TestSwitchToUnknownBranchFails() {
	ctx := context.Background()

	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "v1\n")
	repo.Commit("c1")

	s.Assert().Error(repo.Repo().Switch(ctx, "does-not-exist"))
}

// TestSwitchGuessesRemoteBranch covers the default-guess behaviour: a branch
// that exists only on the remote is checked out by creating a local tracking
// branch for it.
func (s *SwitchSuite) TestSwitchGuessesRemoteBranch() {
	ctx := context.Background()

	producer := gitest.Init(s.T(), s.T().TempDir())
	producer.SetupCommitConfig()
	producer.WriteFileAdd("file", "v1\n")
	producer.Commit("c1")
	start := producer.BranchShowCurrent()
	producer.CheckoutBranch("feature")
	producer.WriteFileAdd("file", "v2\n")
	producer.Commit("c2")
	producer.Checkout(start)

	consumer := gitest.Clone(s.T(), producer.Repo().Path(), s.T().TempDir())

	s.Require().NoError(consumer.Repo().Switch(ctx, "feature"))
	s.Assert().Equal("feature", consumer.BranchShowCurrent())
}
