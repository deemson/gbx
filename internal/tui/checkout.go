package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// checkoutCmd builds a per-repo command that switches to branch off the UI
// goroutine. It is shaped to plug into runOnFiltered: checkoutCmd(branch)
// yields the same func(name, repo) tea.Cmd signature pullCmd already has.
func checkoutCmd(branch string) func(name string, repo git.Repo) tea.Cmd {
	return func(name string, repo git.Repo) tea.Cmd {
		return func() tea.Msg {
			err := repo.Switch(context.Background(), branch)
			if err != nil {
				log.Error().Err(err).Str("name", name).Str("branch", branch).Msg("checkout failed")
			}
			return cmdDoneMsg{name: name, err: err}
		}
	}
}
