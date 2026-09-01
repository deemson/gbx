package git_test

import (
	"context"
	"strings"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type DescribeSuite struct {
	suite.Suite
}

func TestDescribeSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DescribeSuite{})
}

func (s *DescribeSuite) TestCommit() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	description, err := repo.Repo().Describe(context.Background())
	if s.Assert().NoError(err) {
		s.Assert().NotEmpty(description)
		s.Assert().True(strings.HasPrefix(repo.RevParseHead(), description))
	}
}

func (s *DescribeSuite) TestNoCommits() {
	repo := gitest.Init(s.T(), s.T().TempDir())

	description, err := repo.Repo().Describe(context.Background())
	s.Assert().Empty(description)
	s.Assert().ErrorIs(err, git.ErrRepositoryHasNoCommits)
}
