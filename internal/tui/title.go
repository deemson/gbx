package tui

import (
	"io"
	"os/exec"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

const neutralWindowTitle = "gbx"

// actionWindowTitle identifies the repository handed to an interactive child.
// Until all sections finish scanning, multiple roots are conservatively
// qualified because a later result may share the same basename.
func (m model) actionWindowTitle(r repoEntry) string {
	identity := r.name
	if len(m.sections) > 1 && (!m.discoveryComplete() || m.nameInMultipleSections(r.name)) {
		identity = m.sections[r.section].label + "/" + r.name
	}
	identity = stripControls(identity)
	if identity == "" {
		identity = "repo"
	}
	return neutralWindowTitle + " — " + identity
}

func (m model) discoveryComplete() bool {
	for _, section := range m.sections {
		if !section.complete {
			return false
		}
	}
	return true
}

func (m model) nameInMultipleSections(name string) bool {
	section := -1
	for _, repo := range m.repos {
		if repo.name != name {
			continue
		}
		if section >= 0 && section != repo.section {
			return true
		}
		section = repo.section
	}
	return false
}

func stripControls(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// titledExecCommand writes the title only after Bubble Tea has released the
// terminal and supplied the child's streams. A child remains free to replace it.
type titledExecCommand struct {
	cmd    *exec.Cmd
	title  string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func newTitledExecCommand(cmd *exec.Cmd, title string) tea.ExecCommand {
	return &titledExecCommand{cmd: cmd, title: title}
}

func (c *titledExecCommand) SetStdin(r io.Reader)  { c.stdin = r }
func (c *titledExecCommand) SetStdout(w io.Writer) { c.stdout = w }
func (c *titledExecCommand) SetStderr(w io.Writer) { c.stderr = w }

func (c *titledExecCommand) Run() error {
	if c.cmd.Stdin == nil {
		c.cmd.Stdin = c.stdin
	}
	if c.cmd.Stdout == nil {
		c.cmd.Stdout = c.stdout
	}
	if c.cmd.Stderr == nil {
		c.cmd.Stderr = c.stderr
	}
	if c.stdout != nil {
		_, _ = io.WriteString(c.stdout, ansi.SetWindowTitle(c.title))
	}
	return c.cmd.Run()
}
