package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git"
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

func (s *StatusSuite) TestRenamedFile() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFileAdd("file-initial", "data")
	repo.Commit("initial")
	repo.RemovePathAdd("file-initial")
	repo.WriteFileAdd("file-moved", "data")
	repo.WriteFileAdd("file-added", "added")

	branch := repo.BranchShowCurrent()
	commit := repo.RevParseHead()

	status, err := repo.Repo().Status(context.Background())
	if s.Assert().NoError(err) {
		s.Assert().Equal(git.Status{
			Commit: commit,
			Branch: branch,
			Paths: []any{
				git.RegularPathStatus{
					Path: "file-added",
				},
				git.MovedPathStatus{
					Path:     "file-moved",
					OrigPath: "file-initial",
				},
			},
		}, status)
	}
}

func (s *StatusSuite) TestMergeConflicts() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFileAdd("uu-file", "initial data")
	repo.Commit("initial")

	currentBranch := repo.BranchShowCurrent()
	anotherBranch := "branch"

	repo.CheckoutBranch(anotherBranch)
	repo.WriteFileAdd("aa-file", "branch data")
	repo.WriteFileAdd("uu-file", "branch data")
	repo.Commit("branch")

	repo.Checkout(currentBranch)
	repo.WriteFileAdd("aa-file", "main data")
	repo.WriteFileAdd("uu-file", "updated data")
	repo.Commit("update")

	repo.Merge(anotherBranch)

	commit := repo.RevParseHead()

	status, err := repo.Repo().Status(context.Background())
	if s.Assert().NoError(err) {
		s.Assert().Equal(git.Status{
			Commit: commit,
			Branch: currentBranch,
			Paths: []any{
				git.UnmergedPathStatus{
					Path: "aa-file",
				},
				git.UnmergedPathStatus{
					Path: "uu-file",
				},
			},
		}, status)
	}
}
