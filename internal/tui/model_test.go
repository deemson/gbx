package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	appconfig "github.com/deemson/gbx/internal/config"
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

// A scroll key reaches the viewport instead of closing help.
func TestHelpForwardsScrollKeys(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})

	opened, _ := m.Update(keyQuestion)
	scrolled, _ := opened.(model).Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, modeHelp, scrolled.(model).mode) // down scrolls, doesn't close
}

// WithLogPath threads the log file path main.go owns into the model.
func TestWithLogPathSetsField(t *testing.T) {
	cfg := &config{}
	WithLogPath("/state/gbx/gbx-42.log")(cfg)
	require.Equal(t, "/state/gbx/gbx-42.log", cfg.logPath)
}

// The help overlay's header shows the PID and the log path so a failed session
// can be found.
func TestHelpShowsPIDAndLogPath(t *testing.T) {
	m := newModel("x")
	m.pid = 4242
	m.logPath = "/state/gbx/gbx-4242.log"
	m = drive(t, m, tea.WindowSizeMsg{Width: 100, Height: 40}, keyQuestion)

	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "PID: 4242")
	require.Contains(t, out, "/state/gbx/gbx-4242.log")
}

// As the help header narrows, the left log block degrades in rungs while the
// version/PID corner stays pinned: full path → ".../gbx-<pid>.log" → gone.
func TestHelpHeaderLogDegradesButCornerStays(t *testing.T) {
	m := newModel("x")
	m.pid = 4242
	m.logPath = "/state/gbx/gbx-4242.log"

	// Rung 2: too narrow for the full path, wide enough for the abbreviation.
	m.width = 33
	rung2 := ansi.Strip(m.helpHeader())
	require.Contains(t, rung2, "PID: 4242")
	require.Contains(t, rung2, ".../gbx-4242.log")
	require.NotContains(t, rung2, "/state/gbx/gbx-4242.log")

	// Rung 3: too narrow even for the abbreviation — the label and path drop, the
	// corner stays.
	m.width = 20
	rung3 := ansi.Strip(m.helpHeader())
	require.Contains(t, rung3, "PID: 4242")
	require.NotContains(t, rung3, "gbx-4242.log")
	require.NotContains(t, rung3, "Log")
}

// Row 1 carries a dim "<C-f> " hint in front of the filter status (list mode),
// and the always-visible footer shows list-mode action keys.
func TestHeaderHintAndListFooter(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m = drive(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})

	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "<C-f> Filter:")
	require.Contains(t, out, "<C-f> filter") // filter leads the footer now
	require.Contains(t, out, "<r> refresh")
	require.Contains(t, out, "<c> checkout")
	require.Contains(t, out, "<q> quit")
	require.NotContains(t, out, "actions") // enter hint removed from the footer
}

// Opening the filter prompt keeps the "<C-f> " hint and switches the footer to
// the prompt keys; the checkout prompt drops the hint (row 1 isn't the filter).
func TestFooterFollowsPromptMode(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m = drive(t, m, tea.WindowSizeMsg{Width: 100, Height: 40}, keyCtrlF)

	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "<C-f> Filter:")
	require.Contains(t, out, "<enter> apply")
	require.NotContains(t, out, "<r> refresh")

	m = drive(t, newModel("x").addRepo("a", git.Repo{}),
		tea.WindowSizeMsg{Width: 100, Height: 40}, tea.KeyPressMsg{Code: 'c', Text: "c"})
	out = ansi.Strip(m.View().Content)
	require.Contains(t, out, "Checkout:")
	require.NotContains(t, out, "<C-f> Checkout:")
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

func TestCursorMovesAndClamps(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{}).addRepo("c", git.Repo{})
	require.Equal(t, 0, m.cursorIndex()) // starts at the top

	m = drive(t, m, keyDown, keyDown) // 0 -> 2
	require.Equal(t, 2, m.cursorIndex())
	m = drive(t, m, keyDown) // clamps at the last row
	require.Equal(t, 2, m.cursorIndex())
	m = drive(t, m, keyUp, keyUp, keyUp) // clamps at the top
	require.Equal(t, 0, m.cursorIndex())
}

func TestCursorClampsWhenFilterNarrows(t *testing.T) {
	m := newModel("x").addRepo("alpha", git.Repo{}).addRepo("beta", git.Repo{}).addRepo("gamma", git.Repo{})
	m = drive(t, m, keyDown, keyDown) // cursor on the last row
	require.Equal(t, "gamma", m.matched()[m.cursorIndex()].name)

	opened, _ := m.Update(keyCtrlF)
	m = opened.(model)
	m = send(t, m, "alpha") // narrows to a single match
	applied, _ := m.Update(keyEnter)
	m = applied.(model)

	require.Equal(t, 0, m.cursorIndex()) // clamped down to the surviving row
	require.Equal(t, "alpha", m.matched()[m.cursorIndex()].name)
}

func TestCursorBandsOnlyTheCursoredRow(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{}).addRepo("b", git.Repo{})

	lines := strings.Split(m.listContent(), "\n")
	require.Contains(t, lines[0], cursorBandSeq) // cursor starts on the first row
	require.NotContains(t, lines[1], cursorBandSeq)

	m = drive(t, m, keyDown)
	lines = strings.Split(m.listContent(), "\n")
	require.NotContains(t, lines[0], cursorBandSeq)
	require.Contains(t, lines[1], cursorBandSeq) // band follows the cursor down
}

// scrollableModel builds a model of n repos sized to a 5-row scroll window
// (listHeight = height-3-2 = 5), more rows than fit so scrolling is in play.
func scrollableModel(t *testing.T, n int) model {
	t.Helper()
	m := newModel("x")
	for i := 0; i < n; i++ {
		m = m.addRepo(fmt.Sprintf("r%02d", i), git.Repo{})
	}
	return drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 10})
}

func TestCursorMovePastEdgeScrollsWindow(t *testing.T) {
	m := scrollableModel(t, 12) // window 5, margin 2 → cursor locks to the middle row

	m = drive(t, m, ctrlN, ctrlN) // cursor 0 -> 2, still within the top window
	require.Equal(t, 2, m.cursor)
	require.Equal(t, 0, m.top)

	m = drive(t, m, ctrlN) // cursor 3 crosses the bottom margin → window follows
	require.Equal(t, 3, m.cursor)
	require.Equal(t, 1, m.top)
}

func TestHalfPageJumpMovesCursorAndKeepsItVisible(t *testing.T) {
	m := scrollableModel(t, 12) // listHeight 5 → halfPage 2

	m = drive(t, m, ctrlD)
	require.Equal(t, 2, m.cursor) // advanced by a half-page
	m = drive(t, m, ctrlD)
	require.Equal(t, 4, m.cursor)
	require.GreaterOrEqual(t, m.cursor, m.top) // cursor stays in the window
	require.Less(t, m.cursor, m.top+m.listHeight())
}

func TestHalfPageUpAtTopPinsToZero(t *testing.T) {
	m := scrollableModel(t, 12)
	m = drive(t, m, ctrlD, ctrlD, ctrlD) // scroll down first
	require.Positive(t, m.top)

	m = drive(t, m, ctrlU, ctrlU, ctrlU, ctrlU, ctrlU) // overshoot the top
	require.Equal(t, 0, m.cursor)
	require.Equal(t, 0, m.top) // clamped, no wrap
}

func TestFilterNarrowingPullsTopIntoRange(t *testing.T) {
	m := scrollableModel(t, 12)
	for i := 0; i < 12; i++ {
		m = drive(t, m, ctrlN) // cursor to the last row, window scrolled to the end
	}
	require.Equal(t, 7, m.top) // 12 rows - 5-row window

	opened, _ := m.Update(keyCtrlF)
	m = send(t, opened.(model), "^r11") // narrows to a single match
	applied, _ := m.Update(keyEnter)
	m = applied.(model)

	require.Equal(t, 0, m.top) // one match fits the window → scrolled back to top
}

func TestScrollMarkersAppearOnlyWhenContentHidden(t *testing.T) {
	m := scrollableModel(t, 12)

	atTop := ansi.Strip(m.View().Content)
	require.NotContains(t, atTop, "↑")     // nothing hidden above
	require.Contains(t, atTop, "↓ 7 more") // 7 rows below the window

	m = drive(t, m, ctrlD, ctrlD) // cursor 4 → window scrolled into the middle
	middle := ansi.Strip(m.View().Content)
	require.Contains(t, middle, "↑")
	require.Contains(t, middle, "↓")

	for i := 0; i < 12; i++ {
		m = drive(t, m, ctrlN) // to the bottom
	}
	atBottom := ansi.Strip(m.View().Content)
	require.Contains(t, atBottom, "↑")
	require.NotContains(t, atBottom, "↓") // nothing hidden below
}

func TestEnterOpensActionMenuOverCursor(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m.appConfig = appconfig.Config{Actions: []appconfig.Action{{Label: "lazygit", Command: []string{"lazygit"}}}}

	opened, _ := m.Update(keyEnter)
	require.Equal(t, modeActionMenu, opened.(model).mode)
}

func TestActionMenuOverlayRendersTitleAndLabels(t *testing.T) {
	m := newModel("x").addRepo("myrepo", git.Repo{})
	m.appConfig = appconfig.Config{Actions: []appconfig.Action{
		{Label: "lazygit", Command: []string{"lazygit"}},
		{Label: "shell", Command: []string{"sh"}},
	}}
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	opened, _ := m.Update(keyEnter)
	m = opened.(model)
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "Actions — myrepo") // titled with the cursored repo
	require.Contains(t, out, "lazygit")
	require.Contains(t, out, "shell")
}

func TestEnterWithNoMatchesIsNoop(t *testing.T) {
	m := newModel("x") // no repos discovered yet
	opened, _ := m.Update(keyEnter)
	require.Equal(t, modeList, opened.(model).mode)
}

func TestActionMenuEscCloses(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m.mode = modeActionMenu

	closed, _ := m.Update(keyEsc)
	require.Equal(t, modeList, closed.(model).mode)
}

func TestActionDigitFiresAndCloses(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m.appConfig = appconfig.Config{Actions: []appconfig.Action{{Label: "noop", Command: []string{"true"}}}}
	m.mode = modeActionMenu

	fired, cmd := m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	require.Equal(t, modeList, fired.(model).mode) // closes the instant it fires
	require.NotNil(t, cmd)                         // ExecProcess scheduled
}

func TestActionOutOfRangeDigitIgnored(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m.appConfig = appconfig.Config{Actions: []appconfig.Action{{Label: "noop", Command: []string{"true"}}}}
	m.mode = modeActionMenu

	ignored, cmd := m.Update(tea.KeyPressMsg{Code: '2', Text: "2"}) // only one action exists
	require.Equal(t, modeActionMenu, ignored.(model).mode)          // menu stays open
	require.Nil(t, cmd)
}

func TestActionInterpolationErrorSurfacesAsRowError(t *testing.T) {
	m := newModel("x").addRepo("a", git.Repo{})
	m.appConfig = appconfig.Config{Actions: []appconfig.Action{
		{Label: "shell", Command: []string{"{{ env.GBX_DEFINITELY_UNSET }}"}},
	}}
	m.mode = modeActionMenu

	fired, cmd := m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	m = fired.(model)
	require.Equal(t, modeList, m.mode)
	require.NotNil(t, cmd) // nothing launched; the cmd carries the failure back

	m = drive(t, m, cmd()) // feed the cmdDoneMsg it produced
	require.Equal(t, cmdFailed, m.repos[0].cmd)
	require.Contains(t, m.repos[0].summary(), "is not set")
}

func TestEmptyDiffHidesChurn(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})
	m = m.setStatus("r", repoStatus{branch: "main", hasUpstream: true})

	m = m.setDiff("r", lineChanges{}) // +0 -0
	require.NotContains(t, ansi.Strip(m.listContent()), "+0")

	m = m.setDiff("r", lineChanges{added: 2}) // non-empty → shown
	require.Contains(t, ansi.Strip(m.listContent()), "+2 -0")
}

func TestNarrowWidthTruncatesNameAndBranch(t *testing.T) {
	m := newModel("x").addRepo("a-very-long-repo-name", git.Repo{})
	m = m.setStatus("a-very-long-repo-name", repoStatus{branch: "a-very-long-branch-name", hasUpstream: true})
	m = drive(t, m, tea.WindowSizeMsg{Width: 40, Height: 24})

	lines := strings.Split(ansi.Strip(m.listContent()), "\n")
	require.Len(t, lines, 1)
	require.LessOrEqual(t, ansi.StringWidth(lines[0]), 40) // row shrinks to fit the terminal
	require.Contains(t, lines[0], "…")                     // an elastic column was truncated
}

func TestTooNarrowReplacesScreenWithMessage(t *testing.T) {
	m := newModel("x").addRepo("repo", git.Repo{})
	m = drive(t, m, tea.WindowSizeMsg{Width: 18, Height: 24})

	require.True(t, m.tooNarrow())
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "narrow")  // message shown (word-wrapped to fit)
	require.NotContains(t, out, "repo") // the list is gone, not just narrowed
	for _, line := range strings.Split(out, "\n") {
		require.LessOrEqual(t, ansi.StringWidth(line), 18) // every line wrapped within the terminal
	}
}

func TestChipsShortenBeforeCornerDrops(t *testing.T) {
	m := newModel("x").addRepo("repo", git.Repo{})
	m.version = "1.2.3"

	// Narrowest width at which the corner still fits beside the terse chips: in
	// this band the chips are short but the corner is kept — shortening happens
	// one rung before the corner drops.
	m = drive(t, m, tea.WindowSizeMsg{Width: 200, Height: 24})
	w := m.chipsWidth(true) + lipgloss.Width(m.rightBlock()) + cornerGap
	m = drive(t, m, tea.WindowSizeMsg{Width: w, Height: 24})

	require.True(t, m.useShortChips())
	require.True(t, m.showCorner())
	out := ansi.Strip(m.listView())
	require.Contains(t, out, "n+b")              // terse label
	require.Contains(t, out, "gbx 1.2.3")        // corner still shown
	require.NotContains(t, out, "name + branch") // full label dropped first
}

func TestFooterShedsHintsKeepingHelp(t *testing.T) {
	m := newModel("x").addRepo("repo", git.Repo{})
	m = drive(t, m, tea.WindowSizeMsg{Width: 30, Height: 24}) // renderable, but too narrow for the full footer

	footer := ansi.Strip(m.footerLine())
	require.LessOrEqual(t, ansi.StringWidth(footer), 30)
	require.Contains(t, footer, "<?> help") // pinned hint survives the shedding
	require.NotContains(t, footer, "quit")  // a tail hint is dropped whole, not ellipsized
}

func TestHeaderDropsCornerWhenNarrow(t *testing.T) {
	m := newModel("x").addRepo("repo", git.Repo{})
	m.version = "1.2.3"

	wide := drive(t, m, tea.WindowSizeMsg{Width: 120, Height: 24})
	require.True(t, wide.showCorner())
	require.Contains(t, ansi.Strip(wide.listView()), "gbx 1.2.3")

	narrow := drive(t, m, tea.WindowSizeMsg{Width: 30, Height: 24})
	require.False(t, narrow.showCorner())
	require.NotContains(t, ansi.Strip(narrow.listView()), "gbx 1.2.3")
}
