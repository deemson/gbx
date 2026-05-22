package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type RunSuite struct {
	suite.Suite
}

func TestRunSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &RunSuite{})
}

func (s *RunSuite) TestSuccessReturnsStdout() {
	ctx := context.Background()

	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("c1")

	res, err := repo.Repo().Run(ctx, "rev-parse", "HEAD")
	s.Require().NoError(err)
	s.Assert().Equal(0, res.ExitCode)
	s.Assert().Equal(repo.RevParseHead()+"\n", string(res.Stdout))
}

func (s *RunSuite) TestFailureReturnsNonZeroAndErr() {
	ctx := context.Background()

	repo := gitest.Init(s.T(), s.T().TempDir())

	res, err := repo.Repo().Run(ctx, "checkout", "does-not-exist")
	s.Assert().Error(err)
	s.Assert().NotEqual(0, res.ExitCode)
	s.Assert().NotEmpty(res.Stderr)
}
