package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// repoStatus is the display-facing summary of a repo's git state, derived from
// git.Status. Kept small and local to the TUI on purpose — no separate report
// layer.
type repoStatus struct {
	branch      string
	hasUpstream bool
	ahead       int
	behind      int
	changed     int
	conflicts   int
}

func newRepoStatus(s git.Status) repoStatus {
	rs := repoStatus{
		branch:      s.Branch,
		hasUpstream: s.Upstream != "",
		ahead:       s.Ahead,
		behind:      s.Behind,
	}
	for _, p := range s.Paths {
		rs.changed++
		if _, ok := p.(git.ConflictPathStatus); ok {
			rs.conflicts++
		}
	}
	return rs
}

func (rs repoStatus) stateText() string {
	switch {
	case rs.changed == 0:
		return "clean"
	case rs.conflicts > 0:
		return fmt.Sprintf("%d changed, %d conflict", rs.changed, rs.conflicts)
	default:
		return fmt.Sprintf("%d changed", rs.changed)
	}
}

// line renders the status columns shown after the repo name.
func (rs repoStatus) line() string {
	branch := rs.branch
	if !rs.hasUpstream {
		branch += " [no upstream]"
	}
	return fmt.Sprintf("%s  ↑%d ↓%d  %s", branch, rs.ahead, rs.behind, rs.stateText())
}

type statusLoadedMsg struct {
	name   string
	status repoStatus
}

// statusCmd loads one repo's status off the UI goroutine.
func statusCmd(name string, repo git.Repo) tea.Cmd {
	return func() tea.Msg {
		s, err := repo.Status(context.Background())
		if err != nil {
			log.Error().Err(err).Str("name", name).Msg("failed to load status")
			return nil
		}
		return statusLoadedMsg{name: name, status: newRepoStatus(s)}
	}
}
