package tui

import (
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

// drive applies a sequence of messages to the model in order.
func drive(t *testing.T, m model, msgs ...tea.Msg) model {
	t.Helper()
	for _, msg := range msgs {
		u, _ := m.Update(msg)
		m = u.(model)
	}
	return m
}

func TestRepoFoundStartsLoadingCycleAndSpins(t *testing.T) {
	m := drive(t, newModel("x"), repoFoundMsg{name: "r", repo: git.Repo{}})

	require.Equal(t, 3, m.repos[0].loading) // status + diff + branches in flight
	require.True(t, m.spinning)
	require.Equal(t, m.spinner.View(), m.gutterCell(m.repos[0])) // busy row spins
}

func TestLoadCycleCompletesToBlankGutter(t *testing.T) {
	m := drive(t, newModel("x"),
		repoFoundMsg{name: "r", repo: git.Repo{}},
		statusLoadedMsg{name: "r"}, diffLoadedMsg{name: "r"}, branchesLoadedMsg{name: "r"})

	require.Equal(t, 0, m.repos[0].loading)
	require.Empty(t, ansi.Strip(m.gutterCell(m.repos[0]))) // settled, no error → blank
}

func TestLoadFailureSettlesToCross(t *testing.T) {
	m := drive(t, newModel("x"),
		repoFoundMsg{name: "r", repo: git.Repo{}},
		statusLoadedMsg{name: "r"}, diffLoadedMsg{name: "r"},
		loadFailedMsg{name: "r", err: errors.New("read failed")})

	require.Equal(t, 0, m.repos[0].loading)
	require.Equal(t, "✗", ansi.Strip(m.gutterCell(m.repos[0])))
	require.Equal(t, "read failed", m.repos[0].summary())
}

func TestRefreshClearsPriorLoadError(t *testing.T) {
	// First cycle settles with a read failure → ✗.
	m := drive(t, newModel("x"),
		repoFoundMsg{name: "r", repo: git.Repo{}},
		loadFailedMsg{name: "r", err: errors.New("boom")},
		statusLoadedMsg{name: "r"}, diffLoadedMsg{name: "r"})
	require.NotNil(t, m.repos[0].loadErr)

	// `r` dispatches a fresh cycle, which clears the cycle's loadErr up front.
	m = drive(t, m, tea.KeyPressMsg{Code: 'r', Text: "r"})
	require.Equal(t, 3, m.repos[0].loading)
	require.Nil(t, m.repos[0].loadErr)

	// A fully successful cycle leaves the gutter blank.
	m = drive(t, m, statusLoadedMsg{name: "r"}, diffLoadedMsg{name: "r"}, branchesLoadedMsg{name: "r"})
	require.Empty(t, ansi.Strip(m.gutterCell(m.repos[0])))
}

func TestRefreshClearsCommandError(t *testing.T) {
	// A repo settled with a failed command shows ✗ + one-liner.
	m := drive(t, newModel("x"),
		repoFoundMsg{name: "r", repo: git.Repo{}},
		statusLoadedMsg{name: "r"}, diffLoadedMsg{name: "r"}, branchesLoadedMsg{name: "r"},
		cmdDoneMsg{name: "r", err: errors.New("pull boom")},
		statusLoadedMsg{name: "r"}, diffLoadedMsg{name: "r"}, branchesLoadedMsg{name: "r"})
	require.Equal(t, cmdFailed, m.repos[0].cmd)
	require.Equal(t, "pull boom", m.repos[0].summary())

	// `r` wipes the command error; a successful cycle leaves the gutter blank.
	m = drive(t, m, tea.KeyPressMsg{Code: 'r', Text: "r"})
	require.Nil(t, m.repos[0].cmdErr)
	require.Equal(t, cmdNone, m.repos[0].cmd)
	m = drive(t, m, statusLoadedMsg{name: "r"}, diffLoadedMsg{name: "r"}, branchesLoadedMsg{name: "r"})
	require.Empty(t, ansi.Strip(m.gutterCell(m.repos[0])))
	require.Empty(t, m.repos[0].summary())
}

func TestGutterSpinnerWinsWhileBusy(t *testing.T) {
	// A failed command fires a follow-up refresh; the row keeps spinning through
	// it (Q7) and only settles to ✗ once the reads finish.
	m := newModel("x").addRepo("r", git.Repo{})
	m.repos[0].cmd = cmdFailed
	m.repos[0].cmdErr = errors.New("boom")
	m.repos[0].loading = 1

	require.Equal(t, m.spinner.View(), m.gutterCell(m.repos[0]))
}

func TestCmdErrorWinsOneLiner(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})
	m.repos[0].cmdErr = errors.New("cmd boom")
	m.repos[0].loadErr = errors.New("load boom")

	require.Equal(t, "cmd boom", m.repos[0].summary())
}

func TestSpinnerStopsWhenIdle(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})
	m.spinning = true

	updated, cmd := m.Update(spinner.TickMsg{})
	m = updated.(model)

	require.False(t, m.spinning) // nothing busy → tick loop stops
	require.Nil(t, cmd)
}

// These tests drive the model directly, bypassing the terminal renderer (whose
// differential, cursor-positioned output makes in-place state changes invisible
// to raw-output assertions).

// send types each rune of s as a printable key press, returning the resulting
// model.
func send(t *testing.T, m model, s string) model {
	t.Helper()
	for _, r := range s {
		u, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = u.(model)
	}
	return m
}

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

func TestCtrlFOpensFilterPromptPreFilled(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	// Pre-commit a filter so we can verify ctrl+f re-opens with it pre-filled.
	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "foo")
	applied, _ := m.Update(keyEnter)
	m = applied.(model)
	require.Equal(t, modeList, m.mode)
	require.Equal(t, "foo", m.filter)

	reopened, _ := m.Update(keyCtrlF)
	m = reopened.(model)
	require.Equal(t, modeFilterPrompt, m.mode)
	require.Equal(t, "foo", m.prompt.Value()) // pre-filled with committed value
}

func TestCtrlFWhileOpenRevertsAndCloses(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "abc")
	require.Equal(t, "abc", m.prompt.Value())

	// ctrl+f while open discards the draft and returns to list mode, keeping
	// the previously committed filter (empty in this case) unchanged.
	reverted, _ := m.Update(keyCtrlF)
	m = reverted.(model)
	require.Equal(t, modeList, m.mode)
	require.Empty(t, m.filter)
}

func TestEscClearsDraftThenClosesOnEmpty(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})
	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "abc")

	// First ESC just clears the draft, keeping the prompt open.
	cleared, _ := m.Update(keyEsc)
	m = cleared.(model)
	require.Equal(t, modeFilterPrompt, m.mode)
	require.Empty(t, m.prompt.Value())

	// Second ESC, with the draft already empty, reverts and closes.
	closed, _ := m.Update(keyEsc)
	m = closed.(model)
	require.Equal(t, modeList, m.mode)
	require.Empty(t, m.filter)
}

func TestEnterEmptyDraftCommitsAndClearsCommittedFilter(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	// Commit a filter, then re-open and clear via ESC + Enter to commit empty.
	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "foo")
	applied, _ := m.Update(keyEnter)
	m = applied.(model)
	require.Equal(t, "foo", m.filter)

	reopened, _ := m.Update(keyCtrlF)
	m = reopened.(model)
	cleared, _ := m.Update(keyEsc) // draft was "foo" pre-filled → now ""
	m = cleared.(model)
	applied2, _ := m.Update(keyEnter) // commit the empty draft
	m = applied2.(model)
	require.Equal(t, modeList, m.mode)
	require.Empty(t, m.filter) // committed cleared
}

func TestEffectiveFilterUsesDraftWhileEditing(t *testing.T) {
	m := newModel("x").addRepo("api-gateway", git.Repo{}).addRepo("billing", git.Repo{})

	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "api")

	// While the prompt is open, the visible row set tracks the draft live.
	matched := m.matched()
	require.Len(t, matched, 1)
	require.Equal(t, "api-gateway", matched[0].name)
}

func TestEscInListModeIsNoop(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, cmd := m.Update(keyEsc)
	require.Equal(t, modeList, updated.(model).mode)
	require.Nil(t, cmd)
}

func TestPDispatchesPullOnFiltered(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	m = updated.(model)

	require.Equal(t, cmdRunning, m.repos[0].cmd)
	require.NotNil(t, cmd)
}

func TestQQuitsFromList(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.IsType(t, tea.QuitMsg{}, cmd())
}

func TestQInsidePromptTypesIntoDraft(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	m = updated.(model)

	require.Equal(t, modeFilterPrompt, m.mode) // didn't quit
	require.Equal(t, "q", m.prompt.Value())
}

func TestQuestionTogglesHelp(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})

	opened, _ := m.Update(keyQuestion)
	require.Equal(t, modeHelp, opened.(model).mode)

	closedByQuestion, _ := opened.(model).Update(keyQuestion)
	require.Equal(t, modeList, closedByQuestion.(model).mode)

	reopened, _ := closedByQuestion.(model).Update(keyQuestion)
	closedByEsc, _ := reopened.(model).Update(keyEsc)
	require.Equal(t, modeList, closedByEsc.(model).mode)
}

func TestCOpensCheckoutPromptWithSuggestions(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m = m.setBranches("a", []string{"main", "feat"})

	opened, _ := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = opened.(model)

	require.Equal(t, modeCheckoutPrompt, m.mode)
	require.Equal(t, []string{"feat", "main"}, m.suggestions) // visibleBranches, sorted
}

func TestBOpensBranchPromptWithSuggestions(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m = m.setBranches("a", []string{"main", "feat"})

	opened, _ := m.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
	m = opened.(model)

	require.Equal(t, modeBranchPrompt, m.mode)
	// Existing branches are shown as reference (avoid name collisions); Tab can
	// cycle them in, though Enter on an existing name will fail at git layer.
	require.Equal(t, []string{"feat", "main"}, m.suggestions)
}

func TestBranchTabCyclesSuggestions(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m = m.setBranches("a", []string{"main", "feat"})

	opened, _ := m.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
	m = opened.(model)
	tabbed, _ := m.Update(keyTab)
	m = tabbed.(model)

	require.Equal(t, "feat", m.prompt.Value()) // first sorted branch
}

func TestCheckoutTabCyclesBranchSuggestions(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{})
	m = m.setBranches("a", []string{"main", "feat"})
	m = m.setBranches("b", []string{"main", "other"})

	opened, _ := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = opened.(model)
	tabbed, _ := m.Update(keyTab)
	m = tabbed.(model)

	require.Equal(t, "feat", m.prompt.Value()) // first sorted branch
}

func TestCheckoutShiftTabCyclesBackward(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m = m.setBranches("a", []string{"main", "feat"})

	opened, _ := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = opened.(model)
	back, _ := m.Update(keyShiftTab)
	m = back.(model)

	require.Equal(t, "main", m.prompt.Value()) // wraps to last sorted branch
}

func TestCheckoutSuggestionsFilteredByDraft(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m = m.setBranches("a", []string{"main", "feat", "develop"})

	opened, _ := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = opened.(model)
	m = send(t, m, "fe") // fuzzy-match: "feat" survives

	require.Equal(t, []string{"feat"}, m.suggestions)
}

func TestCheckoutEnterRunsOnFiltered(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = opened.(model)
	m = send(t, m, "main")
	applied, cmd := m.Update(keyEnter)
	m = applied.(model)

	require.Equal(t, modeList, m.mode) // closes after run
	require.Equal(t, cmdRunning, m.repos[0].cmd)
	require.NotNil(t, cmd)
}

func TestBranchEnterRunsCheckoutBranch(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	opened, _ := m.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
	m = opened.(model)
	m = send(t, m, "feature")
	applied, cmd := m.Update(keyEnter)
	m = applied.(model)

	require.Equal(t, modeList, m.mode)
	require.Equal(t, cmdRunning, m.repos[0].cmd)
	require.NotNil(t, cmd)
}

func TestFieldToggleChangesFieldButNotLabel(t *testing.T) {
	m := newModel("x")
	require.Equal(t, fieldNameBranch, m.field) // default

	// List mode: ctrl+1/2/3 set the field silently — the modes row in the
	// header reflects it, but the prompt label is no longer field-coupled.
	name, _ := m.Update(ctrl2)
	require.Equal(t, fieldName, name.(model).field)

	// Opening the filter prompt uses the constant "Filter: " label.
	opened, _ := name.(model).Update(keyCtrlF)
	m = opened.(model)
	require.Equal(t, filterLabel, m.prompt.Prompt)

	// Toggling field while the prompt is open updates m.field; the label stays.
	branch, _ := m.Update(ctrl3)
	m = branch.(model)
	require.Equal(t, fieldBranch, m.field)
	require.Equal(t, filterLabel, m.prompt.Prompt)

	back, _ := m.Update(ctrl1)
	m = back.(model)
	require.Equal(t, fieldNameBranch, m.field)
	require.Equal(t, filterLabel, m.prompt.Prompt)
}

func TestBranchFieldFiltersOnBranch(t *testing.T) {
	m := newModel("x").addRepo("api-gateway", git.Repo{}).addRepo("auth-service", git.Repo{})
	m = m.setStatus("api-gateway", repoStatus{branch: "develop"})
	m = m.setStatus("auth-service", repoStatus{branch: "main"})

	field, _ := m.Update(ctrl3) // switch field BEFORE opening the prompt
	m = field.(model)
	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "main")

	matched := m.matched()
	require.Len(t, matched, 1)
	require.Equal(t, "auth-service", matched[0].name)
}

func TestExcludeHidesMatchingRepos(t *testing.T) {
	m := newModel("x").addRepo("api-gateway", git.Repo{}).addRepo("billing", git.Repo{})

	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "!api")

	matched := m.matched()
	require.Len(t, matched, 1)
	require.Equal(t, "billing", matched[0].name)
}

func TestRefreshOnRTargetsFiltered(t *testing.T) {
	m := newModel("x").addRepo("alpha", git.Repo{}).addRepo("beta", git.Repo{})

	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "alpha")
	applied, _ := m.Update(keyEnter)
	m = applied.(model)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	require.NotNil(t, cmd)

	// A filter that matches nothing schedules no work (empty batch is nil).
	opened, _ = m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "zzz") // overrides pre-fill
	applied, _ = m.Update(keyEnter)
	m = applied.(model)
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	require.Nil(t, cmd)
}

func TestColumnWidthsPinnedToFullList(t *testing.T) {
	const short, long = "short", "a-very-long-repo-name"
	m := newModel("x").addRepo(short, git.Repo{}).addRepo(long, git.Repo{})

	// Narrow the visible set to just the short repo via the filter prompt.
	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, short)
	applied, _ := m.Update(keyEnter)
	m = applied.(model)
	require.Len(t, m.matched(), 1)

	// The short name is still padded to the long (filtered-out) name's width,
	// so the next column doesn't shift as the filter narrows.
	pad := strings.Repeat(" ", len(long)-len(short))
	require.Contains(t, ansi.Strip(m.listContent()), short+pad)
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

func TestErrorColumnAlignsAcrossDiffStates(t *testing.T) {
	m := newModel("x").addRepo("aaa", git.Repo{}).addRepo("bbb", git.Repo{})
	// Identical clean status so the name/branch/state columns match across rows;
	// only the diff cell differs (clean → blank vs "+5 -2"). The diff used to be
	// the one variable-width column, floating the error after it.
	st := repoStatus{branch: "main", hasUpstream: true}
	m = m.setStatus("aaa", st).setStatus("bbb", st)
	m = m.setDiff("aaa", lineChanges{})                     // clean → blank diff cell
	m = m.setDiff("bbb", lineChanges{added: 5, deleted: 2}) // "+5 -2"
	m.repos[0].cmdErr = errors.New("boom-aaa")
	m.repos[1].cmdErr = errors.New("boom-bbb")

	lines := strings.Split(ansi.Strip(m.listContent()), "\n")
	require.Len(t, lines, 2)
	idxA := strings.Index(lines[0], "boom-aaa")
	idxB := strings.Index(lines[1], "boom-bbb")
	require.NotEqual(t, -1, idxA)
	require.NotEqual(t, -1, idxB)
	require.Equal(t, idxA, idxB) // error starts at the same column despite differing diff
}

func TestEmptyDiffHidesChurn(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})
	m = m.setStatus("r", repoStatus{branch: "main", hasUpstream: true})

	m = m.setDiff("r", lineChanges{}) // +0 -0
	require.NotContains(t, ansi.Strip(m.listContent()), "+0")

	m = m.setDiff("r", lineChanges{added: 2}) // non-empty → shown
	require.Contains(t, ansi.Strip(m.listContent()), "+2 -0")
}
