package git_test

import (
	"context"
	"testing"

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

func (s *PullSuite) TestFastForward() {
	ctx := context.Background()

	producer := gitest.Init(s.T(), s.T().TempDir())
	producer.SetupCommitConfig()
	producer.WriteFileAdd("file", "v1\n")
	producer.Commit("c1")

	consumer := gitest.Clone(s.T(), producer.Repo().Path(), s.T().TempDir())

	producer.WriteFileAdd("file", "v1\nv2\n")
	producer.Commit("c2")

	s.Require().NoError(consumer.Repo().Pull(ctx))
	s.Assert().Equal(producer.RevParseHead(), consumer.RevParseHead())
}

func (s *PullSuite) TestNoUpstreamFails() {
	ctx := context.Background()

	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("c1")

	s.Assert().Error(repo.Repo().Pull(ctx))
}
