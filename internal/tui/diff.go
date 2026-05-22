package tui

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// lineChanges is the aggregate +added / -deleted across a repo's tracked
// changes vs HEAD, shown in the row's +/- column.
type lineChanges struct {
	added   int
	deleted int
}

func (lc lineChanges) String() string {
	return fmt.Sprintf("+%d -%d", lc.added, lc.deleted)
}

type diffLoadedMsg struct {
	name    string
	changes lineChanges
}

// diffCmd loads one repo's aggregate line changes vs HEAD off the UI goroutine.
// A repo with no commits has no changes vs HEAD and reports +0 -0; any other
// load error is logged and yields no message (the row keeps showing "...").
func diffCmd(name string, repo git.Repo) tea.Cmd {
	return func() tea.Msg {
		d, err := repo.DiffNumStatHead(context.Background())
		if err != nil && !errors.Is(err, git.ErrRepositoryHasNoCommits) {
			log.Error().Err(err).Str("name", name).Msg("failed to load diff")
			return nil
		}
		var lc lineChanges
		for _, p := range d.Paths {
			lc.added += p.AddedLines
			lc.deleted += p.DeletedLines
		}
		return diffLoadedMsg{name: name, changes: lc}
	}
}
