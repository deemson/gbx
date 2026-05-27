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
	require.NotNil(t, cmd) // status + diff + branches auto-refresh scheduled after the command
}

func TestCmdDoneMarksFailed(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(cmdDoneMsg{name: "r", err: errors.New("boom")})
	um := updated.(model)

	require.Equal(t, cmdFailed, um.repos[0].cmd)
}

func TestSummaryFailureShowsError(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(cmdDoneMsg{name: "r", err: git.ErrNotFastForward})
	m = updated.(model)

	require.Equal(t, cmdFailed, m.repos[0].cmd)
	require.Equal(t, git.ErrNotFastForward.Error(), m.repos[0].summary())
}

func TestSummarySuccessIsEmpty(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(cmdDoneMsg{name: "r"})
	m = updated.(model)

	require.Equal(t, cmdOK, m.repos[0].cmd)
	require.Empty(t, m.repos[0].summary())
}

func TestEnterEntersCommandMode(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(keyEnter)
	require.Equal(t, modeCommand, opened.(model).mode)
}

func TestEscFromCommandReturnsToFilterAndClears(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})
	for _, r := range "abc" { // a filter typed in filter mode
		u, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = u.(model)
	}
	require.Equal(t, "abc", m.filter.Value())

	u, _ := m.Update(keyEnter) // → command mode (filter kept)
	m = u.(model)
	require.Equal(t, "abc", m.filter.Value())

	u, _ = m.Update(keyEsc) // → filter mode, cleared
	m = u.(model)
	require.Equal(t, modeFilter, m.mode)
	require.Empty(t, m.filter.Value())
}

func TestEscFromFilterQuits(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	_, cmd := m.Update(keyEsc)
	require.IsType(t, tea.QuitMsg{}, cmd())
}

func TestCommandSubmitRunsOnFiltered(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(keyEnter)
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

	updated, _ := m.Update(keyEnter)
	updated, _ = updated.(model).Update(keyEnter) // submit with empty command
	m = updated.(model)

	require.Equal(t, modeCommand, m.mode)     // empty submit stays in command mode
	require.Equal(t, cmdNone, m.repos[0].cmd) // nothing run
}

func TestUnknownCommandIsNoop(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(keyEnter)
	m = updated.(model)
	for _, r := range "bogus" {
		updated, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(model)
	}
	updated, cmd := m.Update(keyEnter)
	m = updated.(model)

	require.Equal(t, cmdNone, m.repos[0].cmd) // unrecognized command runs nothing
	require.Nil(t, cmd)
	require.Empty(t, m.command.Value()) // line still cleared
}

func TestTabCompletesCommandWord(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(keyEnter)
	m = updated.(model)
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = updated.(model)
	updated, _ = m.Update(keyTab) // first suggestion for "c"
	m = updated.(model)

	require.Equal(t, "checkout", m.command.Value())
}

func TestShiftTabCyclesBackward(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(keyEnter)
	m = updated.(model)
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"}) // matches checkout, fetch
	m = updated.(model)
	updated, _ = m.Update(keyShiftTab) // wraps to the last suggestion
	m = updated.(model)

	require.Equal(t, "fetch", m.command.Value())
}

func TestCheckoutArgSuggestsUnionAndDashB(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{})
	m = m.setBranches("a", []string{"main", "feat"})
	m = m.setBranches("b", []string{"main", "other"})

	// "-b" plus every branch across the visible repos, deduped and sorted.
	require.Equal(t, []string{"-b", "feat", "main", "other"}, m.suggestionsFor("checkout "))
}

func TestCheckoutBranchArgSuggestsUnion(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{})
	m = m.setBranches("a", []string{"main", "feat"})
	m = m.setBranches("b", []string{"main", "other"})

	// every branch across the visible repos, deduped and sorted.
	require.Equal(t, []string{"feat", "main", "other"}, m.suggestionsFor("checkout -b "))
}

func TestSuggestionsFilterByActiveToken(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	require.Equal(t, []string{"fetch"}, m.suggestionsFor("fe")) // only fetch fuzzy-matches "fe"
}

func TestBranchesLoadedPopulatesRow(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})

	updated, _ := m.Update(branchesLoadedMsg{name: "a", branches: []string{"main", "dev"}})
	m = updated.(model)

	require.Equal(t, []string{"main", "dev"}, m.repos[0].branches)
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
	require.Equal(t, modeFilter, closedByEsc.(model).mode)

	closedByCtrlG, _ := opened.(model).Update(ctrlG)
	require.Equal(t, modeFilter, closedByCtrlG.(model).mode)
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
