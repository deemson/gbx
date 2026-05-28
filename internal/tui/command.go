package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
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
