package git_test

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/suite"
)

type DiffNumStatHeadSuite struct {
	suite.Suite
}

func TestDiffNumstatSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DiffNumStatHeadSuite{})
}

func (s *DiffNumStatHeadSuite) TestSome() {
	repo := gitest.Init(s.T(), s.T().TempDir())
	repo.SetupCommitConfig()

	repo.WriteFileAdd("file with added", "line1\nline2\nline3")
	repo.WriteFileAdd("file with deleted", "line1\nline2\nline3")
	repo.WriteFileAdd("file with changed", "line1\nline2\nline3")
	repo.Commit("initial")
	repo.WriteFileAdd("file with added", "line1\nadded\nline2\nline3")
	repo.WriteFileAdd("file with deleted", "line1\nline3")
	repo.WriteFileAdd("file with changed", "line1\nchanged line2\nline3")

	diffNumStat, err := repo.Repo().DiffNumStatHead(context.Background())
	if s.Assert().NoError(err) {
		s.Assert().Equal(git.DiffNumStat{
			Paths: []git.PathDiffNumStat{
				{
					Path:         "file with added",
					AddedLines:   1,
					DeletedLines: 0,
				},
				{
					Path:         "file with changed",
					AddedLines:   1,
					DeletedLines: 1,
				},
				{
					Path:         "file with deleted",
					AddedLines:   0,
					DeletedLines: 1,
				},
			},
		}, diffNumStat)
	}
}
