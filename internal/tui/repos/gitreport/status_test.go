package gitreport_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/deemson/gbx/internal/tui/repos/gitreport"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
)

type StatusSuite struct {
	suite.Suite
}

func TestStatusSuite(t *testing.T) {
	suite.Run(t, &StatusSuite{})
}

func (s *StatusSuite) assert(repo gitest.Repo, expected gitreport.Status) {
	status, err := repo.Repo().Status(context.Background())
	if s.Assert().NoError(err) {
		logger := log.With().Str("test", s.T().Name()).Logger()
		actual := gitreport.NewStatus(logger.WithContext(context.Background()), status)
		s.Assert().Equal(expected, actual)
	}
}

func (s *StatusSuite) TestConflicts() {
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

	s.assert(repo, gitreport.Status{
		Branch:    currentBranch,
		Commit:    commit,
		Conflicts: 4,
	})
}

func (s *StatusSuite) TestDeleted() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFileAdd("deleted-fs", "data")
	repo.WriteFileAdd("deleted-index", "data")
	repo.Commit("initial")

	branch := repo.BranchShowCurrent()
	commit := repo.RevParseHead()

	repo.RemovePath("deleted-fs")
	repo.RemovePathAdd("deleted-index")

	s.assert(repo, gitreport.Status{
		Branch:       branch,
		Commit:       commit,
		DeletedIndex: 1,
		DeletedFS:    1,
	})
}

func (s *StatusSuite) TestModified() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFileAdd("modified-fs", "modified-fs initial")
	repo.WriteFileAdd("modified-index", "modified-index initial")
	repo.Commit("initial")

	branch := repo.BranchShowCurrent()
	commit := repo.RevParseHead()

	repo.WriteFile("modified-fs", "modified-fs changed")
	repo.WriteFileAdd("modified-index", "modified-index changed")

	s.assert(repo, gitreport.Status{
		Branch:        branch,
		Commit:        commit,
		ModifiedIndex: 1,
		ModifiedFS:    1,
	})
}

func (s *StatusSuite) TestAdded() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFile("added-fs", "data")
	repo.WriteFileAdd("added-index", "data")

	branch := repo.BranchShowCurrent()

	s.assert(repo, gitreport.Status{
		Branch:    branch,
		Commit:    git.InitialCommitHash,
		Untracked: 1,
		Added:     1,
	})
}

func (s *StatusSuite) TestMoved() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFileAdd("initial", "data")
	repo.Commit("initial")

	branch := repo.BranchShowCurrent()
	commit := repo.RevParseHead()

	repo.RemovePathAdd("initial")
	repo.WriteFileAdd("moved", "data")

	s.assert(repo, gitreport.Status{
		Branch: branch,
		Commit: commit,
		Moved:  1,
	})
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

	branch := repo.BranchShowCurrent()
	commit := repo.RevParseHead()

	s.assert(repo, gitreport.Status{
		Branch:   branch,
		Commit:   commit,
		Upstream: fmt.Sprintf("origin/%s", anotherRepo.BranchShowCurrent()),
		Ahead:    1,
		Behind:   2,
	})
}
