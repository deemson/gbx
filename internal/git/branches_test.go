package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type BranchesSuite struct {
	suite.Suite
}

func TestBranchesSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &BranchesSuite{})
}

func (s *BranchesSuite) TestLists() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")
	initial := repo.BranchShowCurrent()
	repo.CheckoutBranch("feature")
	repo.CheckoutBranch("other")

	branches, err := repo.Repo().Branches(ctx)
	if s.Assert().NoError(err) {
		s.Assert().ElementsMatch([]string{initial, "feature", "other"}, branches)
	}
}

func (s *BranchesSuite) TestNoCommits() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())

	branches, err := repo.Repo().Branches(ctx)
	if s.Assert().NoError(err) {
		s.Assert().Empty(branches)
	}
}
