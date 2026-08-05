package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type SwitchCreateSuite struct {
	suite.Suite
}

func TestSwitchCreateSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &SwitchCreateSuite{})
}

func (s *SwitchCreateSuite) TestOK() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	err := repo.Repo().SwitchCreate(ctx, "branch")
	if s.Assert().NoError(err) {
		s.Assert().Equal("branch", repo.BranchShowCurrent())
	}
}

func (s *SwitchCreateSuite) TestAlreadyExists() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	err := repo.Repo().SwitchCreate(ctx, "branch")
	if s.Assert().NoError(err) {
		err = repo.Repo().SwitchCreate(ctx, "branch")
		if s.Assert().Error(err) {
			s.Assert().ErrorIs(err, git.ErrBranchAlreadyExists)
		}
	}
}
