package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// pullCmd runs `git pull` on one repo off the UI goroutine.
func pullCmd(name string, repo git.Repo) tea.Cmd {
	return func() tea.Msg {
		err := repo.Pull(context.Background())
		if err != nil {
			log.Error().Err(err).Str("name", name).Msg("pull failed")
		}
		return cmdDoneMsg{name: name, err: err}
	}
}
