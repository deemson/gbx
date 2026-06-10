package git_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type FetchSuite struct {
	suite.Suite
}

func TestFetchSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &FetchSuite{})
}

func (s *FetchSuite) TestOK() {
	ctx := context.Background()
	remoteDir := s.T().TempDir()
	gitest.InitBare(s.T(), remoteDir)

	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.RemoteAdd("origin", remoteDir)

	err := repo.Repo().Fetch(ctx)
	s.Assert().NoError(err)
}

func (s *FetchSuite) TestNoRemote() {
	ctx := context.Background()
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.RemoteAdd("origin", filepath.Join(s.T().TempDir(), "nonexistent"))

	err := repo.Repo().Fetch(ctx)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrNoRemote)
	}
}
