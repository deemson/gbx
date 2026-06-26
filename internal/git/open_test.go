package git_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type OpenSuite struct {
	suite.Suite
}

func TestOpenSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &OpenSuite{})
}

func (s *OpenSuite) TestErrDoesNotExist() {
	dir := s.T().TempDir()
	_, err := git.Open(context.Background(), path.Join(dir, "non-existent"))
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrDoesNotExist)
	}
}

func (s *OpenSuite) TestErrNotDirectory() {
	dir := s.T().TempDir()
	filePath := path.Join(dir, "test")
	s.Require().NoError(os.WriteFile(filePath, []byte("test"), 0755))
	_, err := git.Open(context.Background(), filePath)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrNotDirectory)
	}
}

func (s *OpenSuite) TestErrNotRepository() {
	dir := s.T().TempDir()
	_, err := git.Open(context.Background(), dir)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrNotRepository)
	}
}

func (s *OpenSuite) TestRoot() {
	dir, err := filepath.EvalSymlinks(s.T().TempDir())
	s.Require().NoError(err)
	subDir := path.Join(dir, "sub")

	_ = gitest.Init(s.T(), dir)

	err = os.Mkdir(subDir, 0755)
	s.Require().NoError(err)

	repo, err := git.Open(context.Background(), subDir)
	if s.Assert().NoError(err) {
		s.Assert().Equal(dir, repo.Root())
		s.Assert().Equal(subDir, repo.Path())
	}
}
