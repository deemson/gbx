package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
)

type repoEntry struct {
	name   string
	repo   git.Repo
	status *repoStatus // nil until loaded
	cmd    cmdState
	cmdErr error // last command error, surfaced via drill-in (slice 5)
}

// model is the root TUI model. The filter input is always focused (fzf-style):
// printable keys edit the filter, and every action is a non-printable binding.
type model struct {
	dir    string
	filter textinput.Model
	repos  []repoEntry
	width  int
	height int
}

func newModel(dir string) model {
	filter := textinput.New()
	filter.Prompt = "> "
	filter.Placeholder = "filter repos"
	// Focus here, in the constructor, so the focused state persists into the
	// model the program runs. Calling Focus() in Init() would not, because
	// Init() returns only a Cmd, discarding the mutated model.
	filter.Focus()
	return model{
		dir:    dir,
		filter: filter,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.filter.Focus(), readEntriesCmd(m.dir))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "ctrl+p":
			return m.runOnFiltered(pullCmd)
		}
		// Any other key belongs to the always-focused filter (handled below).
	case entriesLoadedMsg:
		cmds := make([]tea.Cmd, 0, len(msg.entries))
		for _, e := range msg.entries {
			cmds = append(cmds, openRepoCmd(m.dir, e))
		}
		return m, tea.Batch(cmds...)
	case repoFoundMsg:
		return m.addRepo(msg.name, msg.repo), statusCmd(msg.name, msg.repo)
	case statusLoadedMsg:
		return m.setStatus(msg.name, msg.status), nil
	case pullDoneMsg:
		m = m.setCmdResult(msg.name, msg.err)
		if repo, ok := m.repoByName(msg.name); ok {
			return m, statusCmd(msg.name, repo) // auto-refresh after the command
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	return m, cmd
}

// runOnFiltered marks every repo currently matching the filter as running and
// fires cmdFor against each. This is the shared entry point for command
// bindings (pull now; checkout and others later).
func (m model) runOnFiltered(cmdFor func(name string, repo git.Repo) tea.Cmd) (tea.Model, tea.Cmd) {
	pattern := m.filter.Value()
	var cmds []tea.Cmd
	for i := range m.repos {
		if !fuzzyMatch(pattern, m.repos[i].name) {
			continue
		}
		m.repos[i].cmd = cmdRunning
		m.repos[i].cmdErr = nil
		cmds = append(cmds, cmdFor(m.repos[i].name, m.repos[i].repo))
	}
	return m, tea.Batch(cmds...)
}

// addRepo inserts a discovered repo and keeps the list sorted by name.
func (m model) addRepo(name string, repo git.Repo) model {
	m.repos = append(m.repos, repoEntry{name: name, repo: repo})
	sort.Slice(m.repos, func(i, j int) bool {
		return m.repos[i].name < m.repos[j].name
	})
	return m
}

// setStatus attaches a loaded status to the named repo.
func (m model) setStatus(name string, s repoStatus) model {
	for i := range m.repos {
		if m.repos[i].name == name {
			loaded := s
			m.repos[i].status = &loaded
			break
		}
	}
	return m
}

// setCmdResult records the outcome of a command on the named repo.
func (m model) setCmdResult(name string, err error) model {
	for i := range m.repos {
		if m.repos[i].name == name {
			if err != nil {
				m.repos[i].cmd = cmdFailed
				m.repos[i].cmdErr = err
			} else {
				m.repos[i].cmd = cmdOK
			}
			break
		}
	}
	return m
}

func (m model) repoByName(name string) (git.Repo, bool) {
	for i := range m.repos {
		if m.repos[i].name == name {
			return m.repos[i].repo, true
		}
	}
	return git.Repo{}, false
}

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString(m.filter.View())
	b.WriteString("\n\n")

	if len(m.repos) == 0 {
		b.WriteString("no repos")
		return tea.View{Content: b.String(), AltScreen: true}
	}

	pattern := m.filter.Value()
	var matched []repoEntry
	for _, r := range m.repos {
		if fuzzyMatch(pattern, r.name) {
			matched = append(matched, r)
		}
	}
	if len(matched) == 0 {
		b.WriteString("no matches")
		return tea.View{Content: b.String(), AltScreen: true}
	}

	nameWidth := 0
	for _, r := range matched {
		if len(r.name) > nameWidth {
			nameWidth = len(r.name)
		}
	}
	for _, r := range matched {
		fmt.Fprintf(&b, "%-*s  ", nameWidth, r.name)
		if r.status == nil {
			b.WriteString("...")
		} else {
			b.WriteString(r.status.line())
		}
		if g := r.cmd.glyph(); g != "" {
			b.WriteString("  ")
			b.WriteString(g)
		}
		b.WriteString("\n")
	}

	return tea.View{
		Content:   b.String(),
		AltScreen: true,
	}
}
