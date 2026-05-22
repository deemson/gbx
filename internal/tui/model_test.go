package tui

import (
	"errors"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

// These tests drive the model directly, bypassing the terminal renderer (whose
// differential, cursor-positioned output makes in-place state changes invisible
// to raw-output assertions).

func TestPullDoneMarksOKAndSchedulesRefresh(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, cmd := m.Update(pullDoneMsg{name: "r"})
	um := updated.(model)

	require.Equal(t, cmdOK, um.repos[0].cmd)
	require.NotNil(t, cmd) // status auto-refresh scheduled after the command
}

func TestPullDoneMarksFailedAndStoresErr(t *testing.T) {
	m := newModel("x").addRepo("r", git.Repo{})

	updated, _ := m.Update(pullDoneMsg{name: "r", err: errors.New("boom")})
	um := updated.(model)

	require.Equal(t, cmdFailed, um.repos[0].cmd)
	require.Error(t, um.repos[0].cmdErr) // preserved for drill-in
}
