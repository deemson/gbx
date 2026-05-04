package git_test

import (
	"context"
	"fmt"
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

	repo.WriteFile("file untracked", "untracked")
	repo.WriteFileAdd("file moved orig", "moved")
	repo.WriteFileAdd("file modified index", "index initial")
	repo.WriteFileAdd("file modified fs", "fs initial")
	repo.Commit("initial")
	repo.RemovePathAdd("file moved orig")
	repo.WriteFileAdd("file moved", "moved")
	repo.WriteFileAdd("file modified index", "index modified")
	repo.WriteFile("file modified fs", "fs modified")

	branch := repo.BranchShowCurrent()
	commit := repo.RevParseHead()

	status, err := repo.Repo().Status(context.Background())
	if s.Assert().NoError(err) {
		s.Assert().Equal(git.Status{
			Commit: commit,
			Branch: branch,
			Paths: []any{
				git.RegularPathStatus{
					StateIndex: git.NotChangedPathState,
					StateFS:    git.ModifiedPathState,
					Path:       "file modified fs",
				},
				git.RegularPathStatus{
					StateIndex: git.ModifiedPathState,
					StateFS:    git.NotChangedPathState,
					Path:       "file modified index",
				},
				git.MovedPathStatus{
					StateIndex: git.RenamedPathState,
					StateFS:    git.NotChangedPathState,
					Path:       "file moved",
					OrigPath:   "file moved orig",
				},
				git.UntrackedPathStatus{
					Path: "file untracked",
				},
			},
		}, status)
	}
}

func (s *StatusSuite) TestMergeConflicts() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFileAdd("du-file", "initial data")
	repo.WriteFileAdd("ud-file", "initial data")
	repo.WriteFileAdd("uu-file", "initial data")
	repo.Commit("initial")

	currentBranch := repo.BranchShowCurrent()
	anotherBranch := "branch"

	repo.CheckoutBranch(anotherBranch)
	repo.WriteFileAdd("aa-file", "branch data")
	repo.WriteFileAdd("du-file", "branch data")
	repo.RemovePathAdd("ud-file")
	repo.WriteFileAdd("uu-file", "branch data")
	repo.Commit("branch")

	repo.Checkout(currentBranch)
	repo.WriteFileAdd("aa-file", "main data")
	repo.RemovePathAdd("du-file")
	repo.WriteFileAdd("ud-file", "updated data")
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
				git.ConflictPathStatus{
					StateThem: git.AddedPathState,
					StateUs:   git.AddedPathState,
					Path:      "aa-file",
				},
				git.ConflictPathStatus{
					StateThem: git.UpdatedPathState,
					StateUs:   git.DeletedPathState,
					Path:      "du-file",
				},
				git.ConflictPathStatus{
					StateThem: git.DeletedPathState,
					StateUs:   git.UpdatedPathState,
					Path:      "ud-file",
				},
				git.ConflictPathStatus{
					StateThem: git.UpdatedPathState,
					StateUs:   git.UpdatedPathState,
					Path:      "uu-file",
				},
			},
		}, status)
	}
}

func (s *StatusSuite) TestUpstream() {
	remoteRepoDir := s.T().TempDir()
	gitest.InitBare(s.T(), remoteRepoDir)

	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.RemoteAdd("origin", remoteRepoDir)

	anotherRepo := gitest.Init(s.T(), s.T().TempDir())
	anotherRepo.RemoteAdd("origin", remoteRepoDir)
	anotherRepo.SetupCommitConfig()
	anotherRepo.WriteFileAdd("another-repo-file-1", "data")
	anotherRepo.Commit("another repo commit 1")
	anotherRepo.WriteFileAdd("another-repo-file-2", "data")
	anotherRepo.Commit("another repo commit 2")
	anotherRepo.PushSetUpstream("origin", anotherRepo.BranchShowCurrent())

	repo.Fetch()
	repo.SetupCommitConfig()
	repo.WriteFileAdd("repo-file", "data")
	repo.Commit("repo commit")
	repo.BranchSetUpstreamTo("origin", anotherRepo.BranchShowCurrent(), repo.BranchShowCurrent())

	status, err := repo.Repo().Status(context.Background())
	if s.Assert().NoError(err) {
		s.Assert().Equal(git.Status{
			Commit:   repo.RevParseHead(),
			Branch:   repo.BranchShowCurrent(),
			Upstream: fmt.Sprintf("origin/%s", anotherRepo.BranchShowCurrent()),
			Ahead:    1,
			Behind:   2,
		}, status)
	}
}
