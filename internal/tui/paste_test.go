package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

func TestPasteUpdatesFilterPromptWithNormalizedText(t *testing.T) {
	m := newModel("x").addRepo("api-one", git.Repo{})
	m = drive(t, m, keyCtrlF, tea.PasteMsg{Content: "api\tone\n"})

	require.Equal(t, modeFilterPrompt, m.mode)
	require.Equal(t, "api one ", m.prompt.Value())
	require.Equal(t, "api one ", m.effectiveFilter())
}

func TestPasteUpdatesSearchPromptAndJumpsToHit(t *testing.T) {
	m := drive(t, searchModel(), keyCtrlS, tea.PasteMsg{Content: "api\ttwo\n"})

	require.Equal(t, modeSearchPrompt, m.mode)
	require.Equal(t, "api two ", m.prompt.Value())
	require.Equal(t, 1, m.cursor)
}

func TestPasteUpdatesBranchCommandPrompts(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  tea.KeyPressMsg
		mode uiMode
	}{
		{name: "switch", key: tea.KeyPressMsg{Code: 's', Text: "s"}, mode: modeSwitchPrompt},
		{name: "new branch", key: tea.KeyPressMsg{Code: 'b', Text: "b"}, mode: modeBranchPrompt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newModel("x").addRepo("repo", git.Repo{})
			m = drive(t, m, tc.key, tea.PasteMsg{Content: " feature\tname\n "})

			require.Equal(t, tc.mode, m.mode)
			require.Equal(t, " feature name  ", m.prompt.Value())

			m = drive(t, m, keyEnter)
			require.Equal(t, modeList, m.mode)
			require.Equal(t, cmdRunning, m.repos[0].cmd)
		})
	}
}
