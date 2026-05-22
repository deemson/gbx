package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// cmdDoneMsg is the result of a git command finishing on one repo. The model
// records the row's cmdState and auto-refreshes that repo's status and line
// changes. The full output goes to the log, not the UI.
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
		return "⟳"
	case cmdOK:
		return "✓"
	case cmdFailed:
		return "✗"
	default:
		return ""
	}
}

// commandCmd runs an arbitrary git command on one repo off the UI goroutine.
// The full result (exit code, stdout, stderr) is logged; only pass/fail flows
// back to the row via cmdDoneMsg.
func commandCmd(name string, repo git.Repo, args []string) tea.Cmd {
	return func() tea.Msg {
		res, err := repo.Run(context.Background(), args...)
		ev := log.Info()
		if err != nil {
			ev = log.Error().Err(err)
		}
		ev.Str("name", name).
			Strs("args", args).
			Int("exit", res.ExitCode).
			Str("stdout", strings.TrimSpace(string(res.Stdout))).
			Str("stderr", strings.TrimSpace(string(res.Stderr))).
			Msg("command finished")
		return cmdDoneMsg{name: name, err: err}
	}
}
