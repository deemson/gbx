package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type StatusSuite struct {
	suite.Suite
}

func TestStatusSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &StatusSuite{})
}

func (s *StatusSuite) TestBasic() {
	dir := s.T().TempDir()

	ctx := context.Background()
	repo, err := gitest.Init(ctx, dir)
	s.Require().NoError(err)

	repo.WriteFile("untracked-file", []byte("test"))

	_, err = repo.Status(ctx)
	s.Assert().NoError(err)
}
