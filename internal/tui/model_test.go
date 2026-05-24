package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

// These tests drive the model directly, bypassing the terminal renderer (whose
// differential, cursor-positioned output makes in-place state changes invisible
// to raw-output assertions).

func TestCmdDoneMarksOKAndSchedulesRefresh(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, cmd := m.Update(cmdDoneMsg{name: "r"})
	um := updated.(model)

	require.Equal(t, cmdOK, um.repos[0].cmd)
	require.NotNil(t, cmd) // status + diff auto-refresh scheduled after the command
}

func TestCmdDoneMarksFailed(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(cmdDoneMsg{name: "r", err: errors.New("boom")})
	um := updated.(model)

	require.Equal(t, cmdFailed, um.repos[0].cmd)
}

func TestSummaryFailureShowsFirstStderrLine(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(cmdDoneMsg{
		name: "r", args: []string{"pull"}, exit: 1,
		stderr: "error: cannot pull\nhint: stash first\n",
		err:    errors.New("exit 1"),
	})
	m = updated.(model)

	require.Equal(t, cmdFailed, m.repos[0].cmd)
	require.Equal(t, "error: cannot pull", m.repos[0].summary())
}

func TestSummarySuccessShowsLastStdoutLine(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(cmdDoneMsg{
		name: "r", args: []string{"pull"}, exit: 0,
		stdout: "Updating a..b\nFast-forward\nAlready up to date.\n",
	})
	m = updated.(model)

	require.Equal(t, cmdOK, m.repos[0].cmd)
	require.Equal(t, "Already up to date.", m.repos[0].summary())
}

// The output pane is gated on the cursor repo having failed: present for a
// failed repo, gone when the cursor sits on a successful one.
func TestPaneShowsOnlyForFailedCursorRepo(t *testing.T) {
	m := newModel("x").addRepo("bad", git.Repo{}).addRepo("ok", git.Repo{}) // sorted: bad(0), ok(1)

	u, _ := m.Update(cmdDoneMsg{name: "ok", args: []string{"status"}, exit: 0, stdout: "clean\n"})
	m = u.(model)
	u, _ = m.Update(cmdDoneMsg{name: "bad", args: []string{"pull"}, exit: 1, stderr: "fatal: boom\n", err: errors.New("x")})
	m = u.(model)

	name, body := m.cursorOutput() // cursor on "bad"
	require.Equal(t, "bad", name)
	require.Contains(t, body, "fatal: boom")

	moved, _ := m.Update(keyDown) // cursor → "ok"
	m = moved.(model)
	name, _ = m.cursorOutput()
	require.Empty(t, name) // success → no pane
}

func TestPaneRetargetsAsCursorMoves(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{})
	for _, n := range []string{"a", "b"} {
		u, _ := m.Update(cmdDoneMsg{name: n, args: []string{"pull"}, exit: 1, stderr: "err " + n + "\n", err: errors.New("x")})
		m = u.(model)
	}

	require.Equal(t, "a", m.outputName) // cursor 0
	moved, _ := m.Update(keyDown)
	require.Equal(t, "b", moved.(model).outputName) // pane follows the cursor
}

func TestRerunClearsPriorResultAndHidesPane(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})
	u, _ := m.Update(cmdDoneMsg{name: "r", args: []string{"pull"}, exit: 1, stderr: "boom\n", err: errors.New("x")})
	m = u.(model)
	require.NotNil(t, m.repos[0].result)

	u, _ = m.Update(keyTab) // command mode
	m = u.(model)
	for _, r := range "status" {
		u, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = u.(model)
	}
	u, _ = m.Update(keyEnter)
	m = u.(model)

	require.Equal(t, cmdRunning, m.repos[0].cmd)
	require.Nil(t, m.repos[0].result) // prior output dropped
	name, _ := m.cursorOutput()
	require.Empty(t, name) // pane hidden while running
}

func TestPageKeysDoNotEditFilter(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	for _, k := range []tea.KeyPressMsg{keyPgDn, keyPgUp} {
		u, _ := m.Update(k)
		m = u.(model)
	}

	require.Empty(t, m.filter.Value()) // scroll keys are not printable filter input
}

func TestCommandModeTogglesWithTab(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(keyTab)
	require.Equal(t, modeCommand, opened.(model).mode)

	byTab, _ := opened.(model).Update(keyTab)
	require.Equal(t, modeList, byTab.(model).mode)
}

func TestEscQuitsFromCommandMode(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(keyTab)
	stay, cmd := opened.(model).Update(keyEsc)

	require.Equal(t, modeCommand, stay.(model).mode) // esc no longer switches to the filter
	require.IsType(t, tea.QuitMsg{}, cmd())          // it quits instead
}

func TestCursorMovesInCommandMode(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{})
	ctrlJ := tea.KeyPressMsg{Code: 'j', Mod: tea.ModCtrl}

	opened, _ := m.Update(keyTab)
	moved, _ := opened.(model).Update(ctrlJ)

	require.Equal(t, modeCommand, moved.(model).mode) // ctrl+j is not typed into the command line
	require.Equal(t, 1, moved.(model).cursor)         // it moves the repo cursor instead
}

// Arrow keys move the cursor in command mode just like ctrl+j/k and like the
// filter, so navigation is identical in both modes.
func TestArrowKeysMoveCursorInCommandMode(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{})

	opened, _ := m.Update(keyTab)
	down, _ := opened.(model).Update(keyDown)
	require.Equal(t, modeCommand, down.(model).mode)
	require.Equal(t, 1, down.(model).cursor) // down arrow moves the repo cursor

	up, _ := down.(model).Update(keyUp)
	require.Equal(t, 0, up.(model).cursor) // up arrow moves it back
}

func TestCommandSubmitMarksFilteredRunning(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(keyTab)
	m = updated.(model)
	for _, r := range "pull" { // routed to the command input, not the filter
		updated, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(model)
	}
	updated, cmd := m.Update(keyEnter)
	m = updated.(model)

	require.Equal(t, modeCommand, m.mode)        // stays in command mode after submit
	require.Empty(t, m.command.Value())          // the line is cleared for the next command
	require.Equal(t, cmdRunning, m.repos[0].cmd) // filtered repo marked running
	require.NotNil(t, cmd)
}

func TestCommandSubmitEmptyIsNoop(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(keyTab)
	updated, _ = updated.(model).Update(keyEnter) // submit with empty command
	m = updated.(model)

	require.Equal(t, modeCommand, m.mode)     // empty submit stays in command mode
	require.Equal(t, cmdNone, m.repos[0].cmd) // nothing run
}

func TestCursorMovesAndClamps(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{}).addRepo("c", git.Repo{})
	require.Equal(t, 0, m.cursor)

	for i := 0; i < 5; i++ { // more downs than rows
		updated, _ := m.Update(keyDown)
		m = updated.(model)
	}
	require.Equal(t, 2, m.cursor) // clamped to last row

	for i := 0; i < 5; i++ {
		updated, _ := m.Update(keyUp)
		m = updated.(model)
	}
	require.Equal(t, 0, m.cursor) // clamped to first row
}

func TestDiffLoadedPopulatesRow(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})

	updated, _ := m.Update(diffLoadedMsg{name: "a", changes: lineChanges{added: 3, deleted: 1}})
	m = updated.(model)

	require.NotNil(t, m.repos[0].diff)
	require.Equal(t, lineChanges{added: 3, deleted: 1}, *m.repos[0].diff)
}

func TestEmptyDiffHidesChurn(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})
	m = m.setStatus("r", repoStatus{branch: "main", hasUpstream: true})

	m = m.setDiff("r", lineChanges{}) // +0 -0
	require.NotContains(t, ansi.Strip(m.listContent()), "+0")

	m = m.setDiff("r", lineChanges{added: 2}) // non-empty → shown
	require.Contains(t, ansi.Strip(m.listContent()), "+2 -0")
}

func TestColumnWidthsPinnedToFullList(t *testing.T) {
	const short, long = "short", "a-very-long-repo-name"
	m := newModel("x").addRepo(short, git.Repo{}).addRepo(long, git.Repo{})

	for _, r := range short { // filter down to just the short repo
		u, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = u.(model)
	}
	require.Len(t, m.matched(), 1)

	// The short name is still padded to the long (filtered-out) name's width,
	// so the next column doesn't shift as the filter narrows.
	pad := strings.Repeat(" ", len(long)-len(short))
	require.Contains(t, ansi.Strip(m.listContent()), short+pad)
}

func TestHelpTogglesOpenAndClosed(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})

	opened, _ := m.Update(ctrlG)
	require.Equal(t, modeHelp, opened.(model).mode)

	closedByEsc, _ := opened.(model).Update(keyEsc)
	require.Equal(t, modeList, closedByEsc.(model).mode)

	closedByCtrlG, _ := opened.(model).Update(ctrlG)
	require.Equal(t, modeList, closedByCtrlG.(model).mode)
}

func TestFieldSelectsWithCtrl123(t *testing.T) {
	m := newModel("x")
	require.Equal(t, fieldNameBranch, m.field) // default

	name, _ := m.Update(ctrl2)
	require.Equal(t, fieldName, name.(model).field)
	require.Equal(t, "name > ", name.(model).filter.Prompt)

	branch, _ := name.(model).Update(ctrl3)
	require.Equal(t, fieldBranch, branch.(model).field)
	require.Equal(t, "branch > ", branch.(model).filter.Prompt)

	back, _ := branch.(model).Update(ctrl1)
	require.Equal(t, fieldNameBranch, back.(model).field)
	require.Equal(t, "> ", back.(model).filter.Prompt)
}

func TestBranchFieldFiltersOnBranch(t *testing.T) {
	m := newModel("x").addRepo("api-gateway", git.Repo{}).addRepo("auth-service", git.Repo{})
	m = m.setStatus("api-gateway", repoStatus{branch: "develop"})
	m = m.setStatus("auth-service", repoStatus{branch: "main"})

	updated, _ := m.Update(ctrl3) // branch field
	m = updated.(model)
	for _, r := range "main" {
		updated, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(model)
	}

	matched := m.matched()
	require.Len(t, matched, 1)
	require.Equal(t, "auth-service", matched[0].name) // matched by branch, not name
}

func TestExcludeHidesMatchingRepos(t *testing.T) {
	m := newModel("x").addRepo("api-gateway", git.Repo{}).addRepo("billing", git.Repo{})

	for _, r := range "!api" { // DSL negation typed into the always-focused filter
		updated, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(model)
	}

	matched := m.matched()
	require.Len(t, matched, 1)
	require.Equal(t, "billing", matched[0].name) // the "api" match is excluded
}

func TestRefreshTargetsFilteredOnly(t *testing.T) {
	m := newModel("x").addRepo("alpha", git.Repo{}).addRepo("beta", git.Repo{})

	// A filter that matches at least one repo schedules refresh commands.
	for _, r := range "alpha" {
		updated, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(model)
	}
	_, cmd := m.Update(ctrlR)
	require.NotNil(t, cmd)

	// A filter that matches nothing schedules no work (empty batch is nil).
	for _, r := range "zzz" {
		updated, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(model)
	}
	_, cmd = m.Update(ctrlR)
	require.Nil(t, cmd)
}
