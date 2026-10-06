package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

func leftClick(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg(tea.Mouse{X: x, Y: y, Button: tea.MouseLeft})
}

func TestClickSelectsVisibleRepositoryRows(t *testing.T) {
	m := sectionModel().addRepoTo(0, "api", git.Repo{}).addRepoTo(1, "cli", git.Repo{})
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 12})

	// The three-row header puts the first section heading at terminal row 3;
	// its repository is on row 4 and the second section's repo is on row 6.
	updated, cmd := m.Update(leftClick(20, 6))
	require.Nil(t, cmd) // selection never launches an action
	m = updated.(model)
	require.Equal(t, 1, m.cursorIndex())
	require.Equal(t, "cli", m.matched()[m.cursorIndex()].name)

	m = drive(t, m, leftClick(20, 3)) // directory heading
	require.Equal(t, 1, m.cursorIndex())
}

func TestClickAccountsForScrollAndKeepsCursorViewportInvariant(t *testing.T) {
	m := scrollableModel(t, 12)
	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, 2, m.top)

	// Terminal row 5 is the third visible list row: visual line 4, repo 3.
	m = drive(t, m, leftClick(10, 5))
	require.Equal(t, 3, m.cursorIndex())
	require.Equal(t, 2, m.top) // clampView's two-line scroll margin remains in force
}

func TestClickIgnoresNonRepositoryTargetsAndButtons(t *testing.T) {
	m := sectionModel()
	m.sections[0].complete = false // loading status row
	m.sections[1].complete = true  // empty status row
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 12})

	for _, y := range []int{0, 2, 3, 4, 5, 6, 10, 11} { // header, headings, statuses, footer
		m = drive(t, m, leftClick(10, y))
		require.Equal(t, -1, m.cursorIndex(), "row %d", y)
	}

	m = sectionModel().addRepoTo(0, "api", git.Repo{}).addRepoTo(1, "cli", git.Repo{})
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 12})
	m = drive(t, m, tea.MouseClickMsg(tea.Mouse{X: 10, Y: 6, Button: tea.MouseRight}))
	require.Equal(t, 0, m.cursorIndex())
}

func TestClickIgnoresErrorPaneAndNonListModes(t *testing.T) {
	m := newModel("x").addRepo("broken", git.Repo{}).addRepo("clean", git.Repo{})
	m.repos[0].cmdErr = errors.New("broken")
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 14})
	require.Less(t, m.listHeight(), m.baseListHeight())

	m = drive(t, m, leftClick(10, 3+m.listHeight())) // first error-pane row
	require.Equal(t, 0, m.cursorIndex())

	for _, mode := range []uiMode{modeFilterPrompt, modeSearchPrompt, modeSwitchPrompt, modeBranchPrompt, modeHelp, modeActionMenu} {
		next := m
		next.mode = mode
		next = drive(t, next, leftClick(10, 4))
		require.Equal(t, 0, next.cursorIndex(), "mode %d", mode)
		require.Equal(t, mode, next.mode, "mode %d", mode)
	}
}

func TestWheelNavigatesListAndStillScrollsErrorPane(t *testing.T) {
	m := scrollableModel(t, 12)
	m = drive(t, m, tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown}))
	require.Equal(t, 1, m.cursorIndex())
	m = drive(t, m, tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelUp}))
	require.Equal(t, 0, m.cursorIndex())

	m = newModel("x").addRepo("broken", git.Repo{})
	// A multi-line error makes the pane's viewport scrollable.
	m.repos[0].cmdErr = errors.New("line\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline\nline")
	m = drive(t, m, tea.WindowSizeMsg{Width: 60, Height: 14})
	m = drive(t, m, tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown}))
	require.Greater(t, m.errorView.YOffset(), 0)
}

func TestListViewEnablesMouseReporting(t *testing.T) {
	m := newModel("x")
	require.Equal(t, tea.MouseModeCellMotion, m.View().MouseMode)

	m.mode = modeActionMenu
	require.Equal(t, tea.MouseModeNone, m.View().MouseMode)
}
