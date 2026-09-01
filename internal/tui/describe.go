package tui

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

type describeLoadedMsg struct {
	name        string
	description string
}

// describeCmd loads one repo's git description off the UI goroutine. An
// unborn repository has no description and settles successfully; any other
// failure uses the shared row load-error path.
func describeCmd(name string, repo git.Repo) tea.Cmd {
	return func() tea.Msg {
		description, err := repo.Describe(context.Background())
		if err != nil && !errors.Is(err, git.ErrRepositoryHasNoCommits) {
			log.Error().Err(err).Str("name", name).Msg("failed to describe repository")
			return loadFailedMsg{name: name, err: err}
		}
		return describeLoadedMsg{name: name, description: description}
	}
}
