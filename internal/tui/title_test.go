package tui

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

func TestActionWindowTitle(t *testing.T) {
	m := sectionModel().addRepoTo(0, "api", git.Repo{})
	repo := m.repos[0]

	require.Equal(t, "gbx — api", m.actionWindowTitle(repo))

	m.sections[1].complete = false
	require.Equal(t, "gbx — ./services/api", m.actionWindowTitle(repo))

	m.sections[1].complete = true
	m = m.addRepoTo(1, "api", git.Repo{})
	require.Equal(t, "gbx — ./services/api", m.actionWindowTitle(repo))
}

func TestActionWindowTitleStripsControlsAndFallsBack(t *testing.T) {
	m := newModel("x").addRepo("a\x1b]2;bad\x07\n", git.Repo{})
	require.Equal(t, "gbx — a]2;bad", m.actionWindowTitle(m.repos[0]))

	m = newModelWithDirectories([]Directory{{Label: "one\x1b", Path: "one"}, {Label: "two", Path: "two"}})
	m = m.addRepoTo(0, "api\n", git.Repo{})
	require.Equal(t, "gbx — one/api", m.actionWindowTitle(m.repos[0]))

	m = newModel("x").addRepo("\x00\n", git.Repo{})
	require.Equal(t, "gbx — repo", m.actionWindowTitle(m.repos[0]))
}

func TestTitledExecCommandWritesTitleBeforeChildAndForwardsOutput(t *testing.T) {
	cmd := newTitledExecCommand(exec.Command("sh", "-c", "printf child"), "gbx — repo")
	var stdout bytes.Buffer
	cmd.SetStdout(&stdout)
	require.NoError(t, cmd.Run())
	require.Equal(t, ansi.SetWindowTitle("gbx — repo")+"child", stdout.String())
}

func TestTitledExecCommandIgnoresTitleWriteFailure(t *testing.T) {
	cmd := newTitledExecCommand(exec.Command("sh", "-c", "exit 0"), "gbx — repo")
	cmd.SetStdout(failingWriter{})
	require.NoError(t, cmd.Run())

	cmd = newTitledExecCommand(exec.Command("definitely-not-a-gbx-command"), "gbx — repo")
	var stdout bytes.Buffer
	cmd.SetStdout(&stdout)
	require.Error(t, cmd.Run())
	require.Equal(t, ansi.SetWindowTitle("gbx — repo"), stdout.String())
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("no title") }

func TestEveryViewRequestsNeutralWindowTitle(t *testing.T) {
	m := newModel("x")
	require.Equal(t, neutralWindowTitle, m.View().WindowTitle)

	m.mode = modeHelp
	require.Equal(t, neutralWindowTitle, m.View().WindowTitle)

	m.mode = modeActionMenu
	require.Equal(t, neutralWindowTitle, m.View().WindowTitle)

	m.width = 1
	require.Equal(t, neutralWindowTitle, m.View().WindowTitle)
}
