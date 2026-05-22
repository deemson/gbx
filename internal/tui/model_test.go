package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	require.NotNil(t, cmd) // status auto-refresh scheduled after the command
}

func TestCmdDoneMarksFailedAndStoresErr(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(cmdDoneMsg{name: "r", err: errors.New("boom")})
	um := updated.(model)

	require.Equal(t, cmdFailed, um.repos[0].cmd)
	require.Error(t, um.repos[0].cmdErr) // preserved for drill-in
}

func TestCheckoutPromptOpensAndEscCancels(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(ctrlO)
	require.Equal(t, modeBranchPrompt, opened.(model).mode)

	cancelled, _ := opened.(model).Update(keyEsc)
	require.Equal(t, modeList, cancelled.(model).mode)
}

func TestCheckoutSubmitMarksFilteredRunning(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(ctrlO)
	m = updated.(model)
	for _, r := range "main" { // routed to the prompt, not the filter
		updated, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(model)
	}
	updated, cmd := m.Update(keyEnter)
	m = updated.(model)

	require.Equal(t, modeList, m.mode)           // prompt closes on submit
	require.Equal(t, cmdRunning, m.repos[0].cmd) // filtered repo marked running
	require.NotNil(t, cmd)
}

func TestCheckoutSubmitEmptyIsNoop(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(ctrlO)
	updated, _ = updated.(model).Update(keyEnter) // submit with empty branch
	m = updated.(model)

	require.Equal(t, modeList, m.mode)
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

func TestEnterOpensDetailForCursorRepo(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{})

	updated, _ := m.Update(keyDown) // move cursor to "b"
	m = updated.(model)
	updated, cmd := m.Update(keyEnter)
	m = updated.(model)

	require.Equal(t, modeDetail, m.mode)
	require.Equal(t, "b", m.detail.name)
	require.NotNil(t, cmd) // diff load scheduled
}

func TestDetailEscReturnsToList(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})

	updated, _ := m.Update(keyEnter)
	require.Equal(t, modeDetail, updated.(model).mode)

	updated, _ = updated.(model).Update(keyEsc)
	require.Equal(t, modeList, updated.(model).mode)
}

func TestDetailCarriesLastCommandError(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m = m.setCmdResult("a", errors.New("pull boom"))

	updated, _ := m.Update(keyEnter)
	require.Error(t, updated.(model).detail.cmdErr)
}

func TestDetailLoadedPopulatesDiff(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})

	updated, _ := m.Update(keyEnter)
	m = updated.(model)

	diff := git.DiffNumStat{Paths: []git.PathDiffNumStat{{Path: "f.go", AddedLines: 3, DeletedLines: 1}}}
	updated, _ = m.Update(detailLoadedMsg{name: "a", diff: diff})
	m = updated.(model)

	require.True(t, m.detail.loaded)
	require.Len(t, m.detail.diff.Paths, 1)
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
