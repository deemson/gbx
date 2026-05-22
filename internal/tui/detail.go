package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// detailView is the per-repo drill-in: the raw per-file diff vs HEAD plus the
// last command error — neither of which appears in the main row.
type detailView struct {
	name   string
	cmdErr error           // carried from the row (last pull/checkout error)
	diff   git.DiffNumStat // loaded asynchronously
	loaded bool
	err    error // diff load error
}

type detailLoadedMsg struct {
	name string
	diff git.DiffNumStat
	err  error
}

// detailCmd loads a repo's per-file diff vs HEAD off the UI goroutine.
func detailCmd(name string, repo git.Repo) tea.Cmd {
	return func() tea.Msg {
		diff, err := repo.DiffNumStatHead(context.Background())
		if err != nil {
			log.Error().Err(err).Str("name", name).Msg("failed to load detail diff")
		}
		return detailLoadedMsg{name: name, diff: diff, err: err}
	}
}
