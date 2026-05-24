package tui

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

func TestRepoStatusBucketsByType(t *testing.T) {
	rs := newRepoStatus(git.Status{
		Paths: []any{
			// (A,M): added in index, then edited → added wins (index-wins rule).
			git.RegularPathStatus{StateIndex: git.AddedPathState, StateFS: git.ModifiedPathState},
			// (.,M): unstaged modification → modified.
			git.RegularPathStatus{StateIndex: git.NotChangedPathState, StateFS: git.ModifiedPathState},
			// (D,.): staged deletion → deleted.
			git.RegularPathStatus{StateIndex: git.DeletedPathState, StateFS: git.NotChangedPathState},
			git.MovedPathStatus{},     // renamed/copied
			git.UntrackedPathStatus{}, // untracked
			git.ConflictPathStatus{},  // conflict
		},
	})

	require.Equal(t, 1, rs.added)
	require.Equal(t, 1, rs.modified)
	require.Equal(t, 1, rs.deleted)
	require.Equal(t, 1, rs.renamed)
	require.Equal(t, 1, rs.untracked)
	require.Equal(t, 1, rs.conflict)
	require.False(t, rs.clean())
}

// The branch and change-state columns are asserted separately, with color
// stripped, so the test pins glyphs and zero-hiding behavior without coupling
// to the palette's escape codes.
func TestRepoStatusFields(t *testing.T) {
	dirty := newRepoStatus(git.Status{
		Branch:   "main",
		Upstream: "origin/main",
		Ahead:    2,
		Paths: []any{
			git.RegularPathStatus{StateFS: git.ModifiedPathState},
			git.RegularPathStatus{StateFS: git.ModifiedPathState},
			git.RegularPathStatus{StateFS: git.ModifiedPathState},
			git.UntrackedPathStatus{},
			git.UntrackedPathStatus{},
		},
	})
	require.Equal(t, "main ↑2", ansi.Strip(dirty.branchField()))
	require.Equal(t, "~3 …2", ansi.Strip(dirty.stateField()))

	cleanInSync := newRepoStatus(git.Status{Branch: "main", Upstream: "origin/main"})
	require.Equal(t, "main", ansi.Strip(cleanInSync.branchField()))
	require.Equal(t, "✓", ansi.Strip(cleanInSync.stateField()))

	noUpstream := newRepoStatus(git.Status{Branch: "dev"})
	require.Equal(t, "dev ⌀", ansi.Strip(noUpstream.branchField()))
	require.Equal(t, "✓", ansi.Strip(noUpstream.stateField()))

	conflict := newRepoStatus(git.Status{
		Branch:   "feat",
		Upstream: "origin/feat",
		Behind:   1,
		Paths:    []any{git.ConflictPathStatus{}},
	})
	require.Equal(t, "feat ↓1", ansi.Strip(conflict.branchField()))
	require.Equal(t, "‡1", ansi.Strip(conflict.stateField()))
}
