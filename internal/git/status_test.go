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

func (s *StatusSuite) TestSome() {
	ctx := context.Background()
	repo, err := gitest.Init(ctx, s.T().TempDir())
	s.Require().NoError(err)
	s.Require().NoError(repo.SetupCommitConfig(ctx))

	currentBranch, err := repo.Branch(ctx)
	s.Require().NoError(err)

	s.Require().NoError(repo.WriteFile("file", []byte("initial data")))
	s.Require().NoError(repo.Add(ctx, "file"))
	s.Require().NoError(repo.Commit(ctx, "initial"))

	anotherBranch := "branch"
	s.Require().NoError(repo.CheckoutBranch(ctx, anotherBranch))
	s.Require().NoError(repo.WriteFile("file", []byte("branch data")))
	s.Require().NoError(repo.Add(ctx, "file"))
	s.Require().NoError(repo.Commit(ctx, "branch"))

	s.Require().NoError(repo.CheckoutBranch(ctx, currentBranch))
	s.Require().NoError(repo.WriteFile("file", []byte("updated data")))
	s.Require().NoError(repo.Add(ctx, "file"))
	s.Require().NoError(repo.Commit(ctx, "update"))

}

func (s *StatusSuite) TestBasic() {
	s.T().Skip()
	dir := s.T().TempDir()

	ctx := context.Background()
	repo, err := gitest.Init(ctx, dir)
	s.Require().NoError(err)

	repo.WriteFile("untracked-file", []byte("test"))

	_, err = repo.Status(ctx)
	s.Assert().NoError(err)
}

func (s *StatusSuite) TestMergeConflict_Slop() {
	s.T().Skip()
	dir := s.T().TempDir()

	ctx := context.Background()
	repo, err := gitest.Init(ctx, dir)
	s.Require().NoError(err)

	for _, args := range [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"config", "commit.gpgsign", "false"},
	} {
		_, err = repo.Git(ctx, args...)
		s.Require().NoError(err)
	}

	s.Require().NoError(repo.WriteFile("file.txt", []byte("base\n")))
	_, err = repo.Git(ctx, "add", ".")
	s.Require().NoError(err)
	_, err = repo.Git(ctx, "commit", "-m", "base")
	s.Require().NoError(err)
	_, err = repo.Git(ctx, "branch", "-M", "main")
	s.Require().NoError(err)

	_, err = repo.Git(ctx, "checkout", "-b", "feature")
	s.Require().NoError(err)
	s.Require().NoError(repo.WriteFile("file.txt", []byte("feature\n")))
	_, err = repo.Git(ctx, "commit", "-am", "feature")
	s.Require().NoError(err)

	_, err = repo.Git(ctx, "checkout", "main")
	s.Require().NoError(err)
	s.Require().NoError(repo.WriteFile("file.txt", []byte("main\n")))
	_, err = repo.Git(ctx, "commit", "-am", "main")
	s.Require().NoError(err)

	_, err = repo.Git(ctx, "merge", "feature")
	s.Require().Error(err)

	_, err = repo.Status(ctx)
	s.Assert().NoError(err)
}

func (s *StatusSuite) TestSubmodule_Slop() {
	s.T().Skip()
	parentDir := s.T().TempDir()
	subDir := s.T().TempDir()

	ctx := context.Background()

	sub, err := gitest.Init(ctx, subDir)
	s.Require().NoError(err)
	for _, args := range [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"config", "commit.gpgsign", "false"},
	} {
		_, err = sub.Git(ctx, args...)
		s.Require().NoError(err)
	}
	s.Require().NoError(sub.WriteFile("sub.txt", []byte("sub\n")))
	_, err = sub.Git(ctx, "add", ".")
	s.Require().NoError(err)
	_, err = sub.Git(ctx, "commit", "-m", "sub-initial")
	s.Require().NoError(err)

	parent, err := gitest.Init(ctx, parentDir)
	s.Require().NoError(err)
	for _, args := range [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"config", "commit.gpgsign", "false"},
	} {
		_, err = parent.Git(ctx, args...)
		s.Require().NoError(err)
	}
	s.Require().NoError(parent.WriteFile("readme", []byte("parent\n")))
	_, err = parent.Git(ctx, "add", ".")
	s.Require().NoError(err)
	_, err = parent.Git(ctx, "commit", "-m", "parent-initial")
	s.Require().NoError(err)

	_, err = parent.Git(ctx, "-c", "protocol.file.allow=always", "submodule", "add", subDir, "sub")
	s.Require().NoError(err)
	_, err = parent.Git(ctx, "commit", "-m", "add-submodule")
	s.Require().NoError(err)

	s.Require().NoError(parent.WriteFile("sub/new.txt", []byte("dirty\n")))

	_, err = parent.Status(ctx)
	s.Assert().NoError(err)
}
