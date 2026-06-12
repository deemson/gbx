package tui

import (
	"context"
	"errors"
	"os/exec"

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

// cmdState is the result state of the last command run on a repo. It drives the
// left-gutter indicator: cmdRunning spins, cmdFailed settles to ✗, cmdOK/cmdNone
// are blank (success is silent).
type cmdState int

const (
	cmdNone cmdState = iota
	cmdRunning
	cmdOK
	cmdFailed
)

// summary is the one-liner shown after the columns: the typed command error if
// one is set, else the last load cycle's error, else nothing. The command error
// wins so a failed command's reason isn't masked by its follow-up refresh.
func (r repoEntry) summary() string {
	if r.cmdErr != nil {
		return r.cmdErr.Error()
	}
	if r.loadErr != nil {
		return r.loadErr.Error()
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

// runAction suspends the TUI to run a menu action (argv, already interpolated)
// in the cursored repo's directory, handing it the terminal — so an interactive
// tool like lazygit takes over until it exits. The child's exit code is ignored
// (interactive tools exit non-zero for benign reasons); only a launch failure
// — binary missing / not executable — comes back as a row error. Either way the
// repo is refreshed via cmdDoneMsg, since the action may have changed its state.
func runAction(name string, argv []string, dir string) tea.Cmd {
	c := exec.Command(argv[0], argv[1:]...) //nolint:gosec
	c.Dir = dir
	return tea.ExecProcess(c, func(err error) tea.Msg {
		var exitErr *exec.ExitError
		if err != nil && !errors.As(err, &exitErr) {
			return cmdDoneMsg{name: name, err: err}
		}
		return cmdDoneMsg{name: name}
	})
}
