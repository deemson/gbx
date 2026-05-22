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
	require.True(t, opened.(model).branchActive)

	cancelled, _ := opened.(model).Update(keyEsc)
	require.False(t, cancelled.(model).branchActive)
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

	require.False(t, m.branchActive)             // prompt closes on submit
	require.Equal(t, cmdRunning, m.repos[0].cmd) // filtered repo marked running
	require.NotNil(t, cmd)
}

func TestCheckoutSubmitEmptyIsNoop(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(ctrlO)
	updated, _ = updated.(model).Update(keyEnter) // submit with empty branch
	m = updated.(model)

	require.False(t, m.branchActive)
	require.Equal(t, cmdNone, m.repos[0].cmd) // nothing run
}
