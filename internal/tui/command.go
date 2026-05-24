package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

// cmdDoneMsg is the result of a git command finishing on one repo. The model
// records the row's cmdState and the full output, then auto-refreshes that
// repo's status and line changes. The complete output is also logged.
type cmdDoneMsg struct {
	name   string
	args   []string
	exit   int
	stdout string
	stderr string
	err    error
}

// cmdResult is the stored output of the last command run on a repo: the
// condensed one-liner comes from it, and the failure pane shows it in full.
type cmdResult struct {
	args   []string
	exit   int
	stdout string
	stderr string
}

// body is the full stdout/stderr shown in the scrollable failure pane, each
// stream labeled and present only when non-empty.
func (r cmdResult) body() string {
	var parts []string
	if s := strings.TrimRight(r.stdout, "\n"); s != "" {
		parts = append(parts, "stdout:", s)
	}
	if s := strings.TrimRight(r.stderr, "\n"); s != "" {
		if len(parts) > 0 {
			parts = append(parts, "")
		}
		parts = append(parts, "stderr:", s)
	}
	if len(parts) == 0 {
		return "(no output)"
	}
	return strings.Join(parts, "\n")
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

// summary condenses a finished command's output to one line for the row: on
// failure the first stderr line (the error), on success the last stdout line
// (git's summary). Empty while running or before any command.
func (r repoEntry) summary() string {
	if r.result == nil {
		return ""
	}
	switch r.cmd {
	case cmdFailed:
		if s := firstNonEmptyLine(r.result.stderr); s != "" {
			return s
		}
		return firstNonEmptyLine(r.result.stdout)
	case cmdOK:
		if s := lastNonEmptyLine(r.result.stdout); s != "" {
			return s
		}
		return lastNonEmptyLine(r.result.stderr)
	}
	return ""
}

func firstNonEmptyLine(s string) string {
	for _, ln := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(ln); t != "" {
			return t
		}
	}
	return ""
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if t := strings.TrimSpace(lines[i]); t != "" {
			return t
		}
	}
	return ""
}

// commandCmd runs an arbitrary git command on one repo off the UI goroutine.
// The full result (exit code, stdout, stderr) is logged and also carried back
// to the row via cmdDoneMsg for in-app display.
func commandCmd(name string, repo git.Repo, args []string) tea.Cmd {
	return func() tea.Msg {
		res, err := repo.Run(context.Background(), args...)
		stdout := string(res.Stdout)
		stderr := string(res.Stderr)
		ev := log.Info()
		if err != nil {
			ev = log.Error().Err(err)
		}
		ev.Str("name", name).
			Strs("args", args).
			Int("exit", res.ExitCode).
			Str("stdout", strings.TrimSpace(stdout)).
			Str("stderr", strings.TrimSpace(stderr)).
			Msg("command finished")
		return cmdDoneMsg{name: name, args: args, exit: res.ExitCode, stdout: stdout, stderr: stderr, err: err}
	}
}
