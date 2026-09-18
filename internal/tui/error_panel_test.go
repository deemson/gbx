package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
	gitexec "github.com/deemson/gbx/internal/git/exec"
	"github.com/stretchr/testify/require"
)

func detailedError() error {
	return git.NewRunErr(gitexec.Result{
		Args:     []string{"-C", "/tmp/a repo", "pull", "--ff-only"},
		ExitCode: 1,
		Stdout:   []byte("remote output\n"),
		Stderr:   []byte("git's full explanation\nsecond line\n"),
	}, errors.New("exit status 1"), git.ErrNoUpstream)
}

func errorModel(t *testing.T, err error) model {
	t.Helper()
	m := newModel("x").addRepo("broken", git.Repo{})
	m.repos[0].cmdErr = err
	return drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestErrorPanelAppearsAutomaticallyForCursoredError(t *testing.T) {
	m := errorModel(t, detailedError())

	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "Error — broken")
	require.Contains(t, out, "Summary")
	require.Contains(t, out, git.ErrNoUpstream.Error())
	require.Contains(t, out, `git "-C" "/tmp/a repo" "pull" "--ff-only"`)
	require.Contains(t, out, "Exit status")

	details := ansi.Strip(errorDetails(detailedError()))
	require.Contains(t, details, "git's full explanation")
	require.Contains(t, details, "remote output")
}

func TestErrorPanelTracksCursorAndDisappearsOnCleanRow(t *testing.T) {
	m := newModel("x").addRepo("broken", git.Repo{}).addRepo("clean", git.Repo{})
	m.repos[0].cmdErr = errors.New("boom")
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.NotEmpty(t, m.errorViewKey)
	require.Contains(t, ansi.Strip(m.View().Content), "Error — broken")

	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	require.Empty(t, m.errorViewKey)
	require.NotContains(t, ansi.Strip(m.View().Content), "Error — broken")
}

func TestErrorPanelUsesCommandErrorBeforeLoadError(t *testing.T) {
	m := newModel("x").addRepo("broken", git.Repo{})
	m.repos[0].cmdErr = errors.New("command failed")
	m.repos[0].loadErr = errors.New("load failed")
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "command failed")
	require.NotContains(t, out, "load failed")
}

func TestLoadErrorAlsoOpensPanel(t *testing.T) {
	m := newModel("x").addRepo("broken", git.Repo{})
	m.repos[0].loadErr = errors.New("load failed in full")
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	require.Contains(t, ansi.Strip(m.View().Content), "load failed in full")
}

func TestAsyncErrorUnderCursorShowsPanel(t *testing.T) {
	m := newModel("x").addRepo("broken", git.Repo{})
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Empty(t, m.errorViewKey)

	m = drive(t, m, cmdDoneMsg{name: "broken", err: errors.New("arrived later")})
	require.NotEmpty(t, m.errorViewKey)
	require.Contains(t, ansi.Strip(m.View().Content), "arrived later")
}

func TestExplicitModesHideAutomaticErrorPanel(t *testing.T) {
	base := errorModel(t, errors.New("boom"))

	for _, mode := range []uiMode{modeFilterPrompt, modeSearchPrompt, modeSwitchPrompt, modeBranchPrompt, modeHelp, modeActionMenu} {
		m := base
		m.mode = mode
		require.NotContains(t, ansi.Strip(m.View().Content), "Error — broken", "mode %d", mode)
	}
}

func TestErrorPanelScrollsAndResetsForAnotherError(t *testing.T) {
	m := newModel("x").addRepo("first", git.Repo{}).addRepo("second", git.Repo{})
	m.repos[0].cmdErr = errors.New(strings.Repeat("first long line\n", 30))
	m.repos[1].cmdErr = errors.New(strings.Repeat("second long line\n", 30))
	m = drive(t, m, tea.WindowSizeMsg{Width: 60, Height: 14})

	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	require.Greater(t, m.errorView.YOffset(), 0)
	firstKey := m.errorViewKey

	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	require.NotEqual(t, firstKey, m.errorViewKey)
	require.Zero(t, m.errorView.YOffset())
	require.Contains(t, ansi.Strip(m.View().Content), "Error — second")
}

func TestErrorPanePushesListUpAndStaysBetweenListAndFooter(t *testing.T) {
	m := errorModel(t, errors.New(strings.Repeat("very long diagnostic ", 100)))
	spec, ok := m.makeErrorPanelSpec()
	require.True(t, ok)
	require.Equal(t, m.baseListHeight()-spec.paneHeight, m.listHeight())
	require.LessOrEqual(t, spec.paneHeight, errorPaneMaxHeight)

	out := ansi.Strip(m.View().Content)
	rowAt := strings.Index(out, "broken")
	paneAt := strings.Index(out, "Error — broken")
	footerAt := strings.Index(out, "<C-f> filter")
	require.Greater(t, paneAt, rowAt)
	require.Greater(t, footerAt, paneAt)
	require.LessOrEqual(t, lipgloss.Width(m.View().Content), m.width)
	require.LessOrEqual(t, lipgloss.Height(m.View().Content), m.height)
}

func TestSanitizeDiagnosticRemovesTerminalControlsAndExpandsTabs(t *testing.T) {
	got := sanitizeDiagnostic("a\t\x1b[31mred\x1b[0m\r\nnext\x07!")
	require.Equal(t, "a   red\nnext!", got)
	require.NotContains(t, got, "\x1b")
	require.NotContains(t, got, "\x07")
}

func TestGenericErrorPanelUsesCompletePlainText(t *testing.T) {
	m := errorModel(t, errors.New("line one\nline two"))
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "line one")
	require.Contains(t, out, "line two")
	require.NotContains(t, out, "Summary")
}
