package git_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RepoSuite struct {
	suite.Suite
}

func TestRepoSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &RepoSuite{})
}

// func (s *RepoSuite) TestBranch() {
// 	ctx := context.Background()
// 	repo, err := gitest.Init(ctx, s.T().TempDir())
// 	s.Require().NoError(err)
// 	err = repo.CheckoutBranch(ctx, "test")
// 	s.Require().NoError(err)
// 	branch, err := repo.Branch(ctx)
// 	if s.Assert().NoError(err) {
// 		s.Assert().Equal("test", branch)
// 	}
// }
