package gitreport_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/deemson/gbx/internal/tui/repos/gitreport"
	"github.com/stretchr/testify/suite"
)

type LinesChangedSuite struct {
	suite.Suite
}

func TestLinesChangedSuite(t *testing.T) {
	suite.Run(t, &LinesChangedSuite{})
}

func (s *LinesChangedSuite) assert(repo gitest.Repo, expected gitreport.LinesChanged) {
	diffNumStat, err := repo.Repo().DiffNumStatHead(context.Background())
	if s.Assert().NoError(err) {
		actual := gitreport.NewLinesChanged(diffNumStat)
		s.Assert().Equal(expected, actual)
	}
}

func (s *LinesChangedSuite) TestSimple() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFileAdd("file with added", "line1\nline2\nline3")
	repo.WriteFileAdd("file with deleted", "line1\nline2\nline3")
	repo.WriteFileAdd("file with changed", "line1\nline2\nline3")
	repo.Commit("initial")
	repo.WriteFileAdd("file with added", "line1\nadded\nline2\nline3")
	repo.WriteFileAdd("file with deleted", "line1\nline3")
	repo.WriteFileAdd("file with changed", "line1\nchanged line2\nline3")

	s.assert(repo, gitreport.LinesChanged{
		Added:   2,
		Deleted: 2,
	})
}
