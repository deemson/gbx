package tui

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
)

type repoEntry struct {
	name string
	repo git.Repo
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
		}
		// Any other key belongs to the always-focused filter (handled below).
	case entriesLoadedMsg:
		cmds := make([]tea.Cmd, 0, len(msg.entries))
		for _, e := range msg.entries {
			cmds = append(cmds, openRepoCmd(m.dir, e))
		}
		return m, tea.Batch(cmds...)
	case repoFoundMsg:
		return m.addRepo(msg.name, msg.repo), nil
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	return m, cmd
}

// addRepo inserts a discovered repo and keeps the list sorted by name.
func (m model) addRepo(name string, repo git.Repo) model {
	m.repos = append(m.repos, repoEntry{name: name, repo: repo})
	sort.Slice(m.repos, func(i, j int) bool {
		return m.repos[i].name < m.repos[j].name
	})
	return m
}

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString(m.filter.View())
	b.WriteString("\n\n")

	switch {
	case len(m.repos) == 0:
		b.WriteString("no repos")
	default:
		pattern := m.filter.Value()
		matched := 0
		for _, r := range m.repos {
			if fuzzyMatch(pattern, r.name) {
				b.WriteString(r.name)
				b.WriteString("\n")
				matched++
			}
		}
		if matched == 0 {
			b.WriteString("no matches")
		}
	}

	return tea.View{
		Content:   b.String(),
		AltScreen: true,
	}
}
