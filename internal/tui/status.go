package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// repoStatus is the display-facing summary of a repo's git state, derived from
// git.Status. Kept small and local to the TUI on purpose — no separate report
// layer. Changed files are bucketed by type, one bucket per file.
type repoStatus struct {
	branch      string
	hasUpstream bool
	ahead       int
	behind      int
	modified    int
	added       int
	deleted     int
	renamed     int
	untracked   int
	conflict    int
}

func newRepoStatus(s git.Status) repoStatus {
	rs := repoStatus{
		branch:      s.Branch,
		hasUpstream: s.Upstream != "",
		ahead:       s.Ahead,
		behind:      s.Behind,
	}
	for _, p := range s.Paths {
		switch p := p.(type) {
		case git.UntrackedPathStatus:
			rs.untracked++
		case git.ConflictPathStatus:
			rs.conflict++
		case git.MovedPathStatus:
			rs.renamed++ // a rename/copy is its own entry type
		case git.RegularPathStatus:
			switch significantState(p.StateIndex, p.StateFS) {
			case git.ModifiedPathState:
				rs.modified++
			case git.AddedPathState:
				rs.added++
			case git.DeletedPathState:
				rs.deleted++
			}
		}
	}
	return rs
}

// significantState applies the index-wins rule: a file is bucketed by its index
// (X) state when that side records a change, otherwise by its worktree (Y)
// state. So (A,M) counts as added, (.,M) as modified.
func significantState(index, fs git.PathState) git.PathState {
	if index != git.NotChangedPathState && index != git.UnknownPathState {
		return index
	}
	return fs
}

func (rs repoStatus) clean() bool {
	return rs.modified+rs.added+rs.deleted+rs.renamed+rs.untracked+rs.conflict == 0
}

// branchField is the branch name plus its sync state: a dim ⌀ when the branch
// has no upstream, otherwise the non-zero ahead/behind arrows (nothing when in
// sync). The branch name itself keeps the default foreground.
func (rs repoStatus) branchField() string {
	if !rs.hasUpstream {
		return rs.branch + " " + colorDim.Render("⌀")
	}
	if sync := rs.sync(); sync != "" {
		return rs.branch + " " + colorCyan.Render(sync)
	}
	return rs.branch
}

// sync is the ahead/behind arrows with zero sides hidden, empty when in sync.
func (rs repoStatus) sync() string {
	s := ""
	if rs.ahead > 0 {
		s += fmt.Sprintf("↑%d", rs.ahead)
	}
	if rs.behind > 0 {
		s += fmt.Sprintf("↓%d", rs.behind)
	}
	return s
}

// stateField is blank when the tree is clean (success is silent), otherwise the
// non-empty change buckets, each a colored glyph+count, in a stable order.
func (rs repoStatus) stateField() string {
	if rs.clean() {
		return ""
	}
	var segs []string
	if rs.modified > 0 {
		segs = append(segs, colorYellow.Render(fmt.Sprintf("~%d", rs.modified)))
	}
	if rs.added > 0 {
		segs = append(segs, colorGreen.Render(fmt.Sprintf("✚%d", rs.added)))
	}
	if rs.deleted > 0 {
		segs = append(segs, colorRed.Render(fmt.Sprintf("✖%d", rs.deleted)))
	}
	if rs.renamed > 0 {
		segs = append(segs, colorMagenta.Render(fmt.Sprintf("»%d", rs.renamed)))
	}
	if rs.untracked > 0 {
		segs = append(segs, colorDim.Render(fmt.Sprintf("…%d", rs.untracked)))
	}
	if rs.conflict > 0 {
		segs = append(segs, colorBrightRed.Render(fmt.Sprintf("‡%d", rs.conflict)))
	}
	return strings.Join(segs, " ")
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
			return loadFailedMsg{name: name, err: err}
		}
		return statusLoadedMsg{name: name, status: newRepoStatus(s)}
	}
}
