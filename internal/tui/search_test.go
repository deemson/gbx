package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

// searchModel builds a three-repo model whose branches let name and branch
// searches diverge: "api" matches the two api-* names, but on the branch field
// only "zeta" (branch "api-branch") matches. Sorted order is api-one(0),
// api-two(1), zeta(2).
func searchModel() model {
	m := newModel("x").
		addRepo("api-one", git.Repo{}).
		addRepo("api-two", git.Repo{}).
		addRepo("zeta", git.Repo{})
	m = m.setStatus("api-one", repoStatus{branch: "main"})
	m = m.setStatus("api-two", repoStatus{branch: "main"})
	m = m.setStatus("zeta", repoStatus{branch: "api-branch"})
	return m
}

func TestCtrlSOpensSearchPromptSnapshottingCursor(t *testing.T) {
	m := drive(t, searchModel(), keyDown) // cursor on row 1

	opened, _ := m.Update(keyCtrlS)
	m = opened.(model)

	require.Equal(t, modeSearchPrompt, m.mode)
	require.Equal(t, 1, m.origCursor) // snapshotted for cancel
	require.Empty(t, m.prompt.Value())
}

func TestCtrlSNoopWhenNothingMatched(t *testing.T) {
	m := newModel("x") // no repos → matched() empty

	opened, _ := m.Update(keyCtrlS)
	require.Equal(t, modeList, opened.(model).mode)
}

func TestSearchJumpsCursorToFirstHitWithoutNarrowing(t *testing.T) {
	m := drive(t, searchModel(), keyCtrlS)
	m = send(t, m, "zeta") // matches only row 2

	require.Equal(t, 2, m.cursor)
	require.Len(t, m.matched(), 3) // search never narrows the visible set
}

func TestSearchNoHitLeavesCursorPut(t *testing.T) {
	m := drive(t, searchModel(), keyDown, keyCtrlS) // cursor on row 1
	m = send(t, m, "qqq")                           // no name or branch has a "q"

	require.Equal(t, 1, m.cursor) // unchanged
}

func TestSearchCtrlNCtrlPWalkMatchesWithWrap(t *testing.T) {
	m := drive(t, searchModel(), keyCtrlS, ctrl2) // name field only
	m = send(t, m, "api")                         // hits the two api-* names (rows 0, 1), not zeta
	require.Equal(t, 0, m.cursor)

	m = drive(t, m, ctrlN)
	require.Equal(t, 1, m.cursor)
	m = drive(t, m, ctrlN) // past the last hit → wraps to the first
	require.Equal(t, 0, m.cursor)
	m = drive(t, m, ctrlP) // before the first hit → wraps to the last
	require.Equal(t, 1, m.cursor)
	m = drive(t, m, ctrlP)
	require.Equal(t, 0, m.cursor)
}

func TestSearchArrowsWalkMatches(t *testing.T) {
	m := drive(t, searchModel(), keyCtrlS, ctrl2) // name field only
	m = send(t, m, "api")                         // hits rows 0, 1
	require.Equal(t, 0, m.cursor)

	m = drive(t, m, keyDown) // next match
	require.Equal(t, 1, m.cursor)
	m = drive(t, m, keyDown) // wraps to the first
	require.Equal(t, 0, m.cursor)
	m = drive(t, m, keyUp) // wraps to the last
	require.Equal(t, 1, m.cursor)
}

func TestSearchFieldChangeRejumps(t *testing.T) {
	m := drive(t, searchModel(), keyCtrlS)
	m = send(t, m, "api") // name+branch: first hit is row 0
	require.Equal(t, 0, m.cursor)

	m = drive(t, m, ctrl3) // branch field: only "zeta" (branch api-branch) matches
	require.Equal(t, 2, m.cursor)
	require.Equal(t, fieldBranch, m.field)
}

func TestSearchEnterKeepsLandingSpot(t *testing.T) {
	m := drive(t, searchModel(), keyCtrlS)
	m = send(t, m, "zeta")
	require.Equal(t, 2, m.cursor)

	committed, _ := m.Update(keyEnter)
	m = committed.(model)

	require.Equal(t, modeList, m.mode)
	require.Equal(t, 2, m.cursor) // kept where search landed
}

func TestSearchEscRestoresCursor(t *testing.T) {
	m := drive(t, searchModel(), keyCtrlS) // opens at cursor 0
	m = send(t, m, "zeta")
	require.Equal(t, 2, m.cursor)

	cancelled, _ := m.Update(keyEsc)
	m = cancelled.(model)

	require.Equal(t, modeList, m.mode)
	require.Equal(t, 0, m.cursor) // restored to where search opened
}

func TestSearchCtrlSRestoresCursorAndCloses(t *testing.T) {
	m := drive(t, searchModel(), keyDown, keyCtrlS) // opens at cursor 1
	m = send(t, m, "zeta")
	require.Equal(t, 2, m.cursor)

	cancelled, _ := m.Update(keyCtrlS) // re-press cancels
	m = cancelled.(model)

	require.Equal(t, modeList, m.mode)
	require.Equal(t, 1, m.cursor)
}

// Search moves the cursor without routing through the filter: the committed
// filter is unchanged while searching, and the highlight tracks the search
// draft instead.
func TestSearchDoesNotTouchTheFilter(t *testing.T) {
	m := drive(t, searchModel(), keyCtrlS)
	m = send(t, m, "api")

	require.Empty(t, m.effectiveFilter())        // committed filter untouched
	require.Equal(t, "api", m.highlightSource()) // underline follows the search draft
}

// While searching, rows the cursor can't reach (non-matches) render flat grey,
// and matching rows keep their colors. An empty draft dims nothing.
func TestSearchDimsNonMatchingRows(t *testing.T) {
	m := newModel("x").addRepo("alpha", git.Repo{}).addRepo("beta", git.Repo{})
	clean := repoStatus{branch: "main", hasUpstream: true}
	m = m.setStatus("alpha", clean).setStatus("beta", clean)
	m = m.setDiff("alpha", lineChanges{}).setDiff("beta", lineChanges{})
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	probe := colorDim.Render("X")
	dimOpen := probe[:strings.Index(probe, "X")] // the "enter dim" escape

	// Empty draft: nothing is dimmed yet.
	opened, _ := m.Update(keyCtrlS)
	m = opened.(model)
	for line := range strings.SplitSeq(m.listContent(), "\n") {
		require.NotContains(t, line, dimOpen)
	}

	// "beta" matches only row 1; row 0 (alpha) greys out, row 1 keeps its colors.
	m = send(t, m, "beta")
	lines := strings.Split(m.listContent(), "\n")
	require.Contains(t, lines[0], dimOpen)    // non-matching row dimmed
	require.NotContains(t, lines[1], dimOpen) // matching row not dimmed
}

// Chrome: list mode carries a "<C-s> search" footer hint; opening search shows
// the "<C-s> Search:" header row and switches the footer to the search keys.
func TestSearchChrome(t *testing.T) {
	base := drive(t, searchModel(), tea.WindowSizeMsg{Width: 120, Height: 40})

	list := ansi.Strip(base.View().Content)
	require.Contains(t, list, "<C-s> search") // list footer hint

	opened, _ := base.Update(keyCtrlS)
	out := ansi.Strip(opened.(model).View().Content)
	require.Contains(t, out, "<C-s> Search:") // header row 1
	require.Contains(t, out, "keep")          // footer switched to search keys
	require.Contains(t, out, "next/prev")
	require.NotContains(t, out, "<r> refresh") // list-mode hints gone
}

// End-to-end smoke test that ctrl+s routes through the real program and opens
// the search prompt (the terminal-level flow-control check is manual).
func TestCtrlSOpensSearchEndToEnd(t *testing.T) {
	dir := t.TempDir()
	mkRepo(t, dir, "myrepo")

	tp := runTestProgram(t, dir)
	tp.waitForContent("myrepo")

	tp.sendKey(keyCtrlS)
	tp.waitForContent("Search:")
}
