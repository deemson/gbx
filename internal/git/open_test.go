package git_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/suite"
)

type OpenSuite struct {
	suite.Suite
}

func TestOpenSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &OpenSuite{})
}

func (s *OpenSuite) TestOpen() {
	dir := s.T().TempDir()
	_, err := git.Open(context.Background(), path.Join(dir, "non-existent"))
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err,  git.ErrDoesNotExist)
	}
}

func (s *OpenSuite) TestFile() {
	dir := s.T().TempDir()
	filePath := path.Join(dir, "test")
	s.Require().NoError(os.WriteFile(filePath, []byte("test"), 0755))
	_, err := git.Open(context.Background(), filePath)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err, git.ErrNotDirectory)
	}
}

func (s *OpenSuite) TestNonGitDir() {
	dir := s.T().TempDir()
	_, err := git.Open(context.Background(), dir)
	if s.Assert().Error(err) {
		s.Assert().ErrorIs(err,  git.ErrNotRepository)
	}
}
