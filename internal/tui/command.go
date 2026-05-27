package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// cmdDoneMsg is the result of a command finishing on one repo. The model records
// the row's cmdState and the typed error (nil on success), then auto-refreshes
// that repo's status, line changes, and branches. err is also logged.
type cmdDoneMsg struct {
	name string
	err  error
}

// cmdState is the result state of the last command run on a repo, rendered as a
// glyph in the row's result cell.
type cmdState int

const (
	cmdNone cmdState = iota
	cmdRunning
	cmdOK
	cmdFailed
)

func (c cmdState) glyph() string {
	switch c {
	case cmdRunning:
		return colorYellow.Render("⟳")
	case cmdOK:
		return colorGreen.Render("✓")
	case cmdFailed:
		return colorRed.Render("✗")
	default:
		return ""
	}
}

// summary is the one-liner shown after a command: the typed error on failure
// (the whole point of the strict command set — errors are known), nothing on
// success or while running.
func (r repoEntry) summary() string {
	if r.cmd == cmdFailed && r.cmdErr != nil {
		return r.cmdErr.Error()
	}
	return ""
}

// parseCommand maps a parsed command line to the action run against each repo,
// or reports !ok when the line is not one of the four supported commands:
//
//	checkout <ref>   checkout -b <name>   fetch   pull
func parseCommand(fields []string) (func(name string, repo git.Repo) tea.Cmd, bool) {
	switch {
	case len(fields) == 1 && fields[0] == "fetch":
		return func(name string, repo git.Repo) tea.Cmd {
			return runCmd(name, "fetch", repo.Fetch)
		}, true
	case len(fields) == 1 && fields[0] == "pull":
		return func(name string, repo git.Repo) tea.Cmd {
			return runCmd(name, "pull", repo.Pull)
		}, true
	case len(fields) == 2 && fields[0] == "checkout" && fields[1] != "-b":
		ref := fields[1]
		return func(name string, repo git.Repo) tea.Cmd {
			return runCmd(name, "checkout", func(ctx context.Context) error { return repo.Checkout(ctx, ref) })
		}, true
	case len(fields) == 3 && fields[0] == "checkout" && fields[1] == "-b":
		branch := fields[2]
		return func(name string, repo git.Repo) tea.Cmd {
			return runCmd(name, "checkout -b", func(ctx context.Context) error { return repo.CheckoutBranch(ctx, branch) })
		}, true
	default:
		return nil, false
	}
}

// runCmd runs one typed git method on one repo off the UI goroutine, logging and
// carrying back the typed error via cmdDoneMsg for the row glyph and one-liner.
func runCmd(name, label string, run func(context.Context) error) tea.Cmd {
	return func() tea.Msg {
		err := run(context.Background())
		ev := log.Info()
		if err != nil {
			ev = log.Error().Err(err)
		}
		ev.Str("name", name).Str("command", label).Msg("command finished")
		return cmdDoneMsg{name: name, err: err}
	}
}
