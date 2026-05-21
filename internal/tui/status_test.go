package tui

import (
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

func TestRepoStatusLine(t *testing.T) {
	withUpstream := newRepoStatus(git.Status{
		Branch:   "main",
		Upstream: "origin/main",
		Ahead:    2,
		Behind:   1,
		Paths:    []any{git.RegularPathStatus{}, git.ConflictPathStatus{}},
	})
	require.Equal(t, "main  ↑2 ↓1  2 changed, 1 conflict", withUpstream.line())

	clean := newRepoStatus(git.Status{Branch: "main"})
	require.Equal(t, "main [no upstream]  ↑0 ↓0  clean", clean.line())

	dirty := newRepoStatus(git.Status{
		Branch:   "dev",
		Upstream: "origin/dev",
		Paths:    []any{git.UntrackedPathStatus{}, git.RegularPathStatus{}, git.RegularPathStatus{}},
	})
	require.Equal(t, "dev  ↑0 ↓0  3 changed", dirty.line())
}
