package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

type branchesLoadedMsg struct {
	name     string
	branches []string
}

// branchesCmd loads one repo's local branch names off the UI goroutine. They
// feed checkout autocomplete; a load error is logged and yields no message (the
// repo just contributes nothing to the suggestions).
func branchesCmd(name string, repo git.Repo) tea.Cmd {
	return func() tea.Msg {
		branches, err := repo.Branches(context.Background())
		if err != nil {
			log.Error().Err(err).Str("name", name).Msg("failed to load branches")
			return nil
		}
		return branchesLoadedMsg{name: name, branches: branches}
	}
}
