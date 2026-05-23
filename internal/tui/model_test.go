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
	require.NotNil(t, cmd) // status + diff auto-refresh scheduled after the command
}

func TestCmdDoneMarksFailed(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(cmdDoneMsg{name: "r", err: errors.New("boom")})
	um := updated.(model)

	require.Equal(t, cmdFailed, um.repos[0].cmd)
}

func TestCommandModeTogglesWithTabAndEsc(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(keyTab)
	require.Equal(t, modeCommand, opened.(model).mode)

	byTab, _ := opened.(model).Update(keyTab)
	require.Equal(t, modeList, byTab.(model).mode)

	reopened, _ := m.Update(keyTab)
	byEsc, _ := reopened.(model).Update(keyEsc)
	require.Equal(t, modeList, byEsc.(model).mode)
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

	require.Equal(t, modeList, m.mode)           // command mode closes on submit
	require.Equal(t, cmdRunning, m.repos[0].cmd) // filtered repo marked running
	require.NotNil(t, cmd)
}

func TestCommandSubmitEmptyIsNoop(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(keyTab)
	updated, _ = updated.(model).Update(keyEnter) // submit with empty command
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

func TestDiffLoadedPopulatesRow(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})

	updated, _ := m.Update(diffLoadedMsg{name: "a", changes: lineChanges{added: 3, deleted: 1}})
	m = updated.(model)

	require.NotNil(t, m.repos[0].diff)
	require.Equal(t, "+3 -1", m.repos[0].diff.String())
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

func TestPolarityTogglesWithCtrl1(t *testing.T) {
	m := newModel("x")
	require.Equal(t, polarityInclude, m.polarity) // default

	toggled, _ := m.Update(ctrl1)
	require.Equal(t, polarityExclude, toggled.(model).polarity)
	require.Equal(t, "! ", toggled.(model).filter.Prompt)

	back, _ := toggled.(model).Update(ctrl1)
	require.Equal(t, polarityInclude, back.(model).polarity)
	require.Equal(t, "> ", back.(model).filter.Prompt)
}

func TestFieldSelectsWithCtrl234(t *testing.T) {
	m := newModel("x")
	require.Equal(t, fieldNameBranch, m.field) // default

	name, _ := m.Update(ctrl3)
	require.Equal(t, fieldName, name.(model).field)
	require.Equal(t, "name > ", name.(model).filter.Prompt)

	branch, _ := name.(model).Update(ctrl4)
	require.Equal(t, fieldBranch, branch.(model).field)
	require.Equal(t, "branch > ", branch.(model).filter.Prompt)

	back, _ := branch.(model).Update(ctrl2)
	require.Equal(t, fieldNameBranch, back.(model).field)
	require.Equal(t, "> ", back.(model).filter.Prompt)
}

func TestAxesAreOrthogonal(t *testing.T) {
	m := newModel("x")

	ex, _ := m.Update(ctrl1)          // exclude
	br, _ := ex.(model).Update(ctrl4) // branch field, leaving polarity alone
	bm := br.(model)

	require.Equal(t, polarityExclude, bm.polarity)
	require.Equal(t, fieldBranch, bm.field)
	require.Equal(t, "branch ! ", bm.filter.Prompt)
}

func TestBranchFieldFiltersOnBranch(t *testing.T) {
	m := newModel("x").addRepo("api-gateway", git.Repo{}).addRepo("auth-service", git.Repo{})
	m = m.setStatus("api-gateway", repoStatus{branch: "develop"})
	m = m.setStatus("auth-service", repoStatus{branch: "main"})

	updated, _ := m.Update(ctrl4) // branch field
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

	updated, _ := m.Update(ctrl1) // exclude
	m = updated.(model)
	for _, r := range "api" {
		updated, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
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
