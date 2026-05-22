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
	// branch is the transient checkout prompt, shown and given key focus only
	// while branchActive. The filter is blurred for its duration.
	branch       textinput.Model
	branchActive bool
	repos        []repoEntry
	width        int
	height       int
}

func newModel(dir string) model {
	filter := textinput.New()
	filter.Prompt = "> "
	filter.Placeholder = "filter repos"
	// Focus here, in the constructor, so the focused state persists into the
	// model the program runs. Calling Focus() in Init() would not, because
	// Init() returns only a Cmd, discarding the mutated model.
	filter.Focus()
	branch := textinput.New()
	branch.Prompt = "branch: "
	branch.Placeholder = "switch to branch"
	return model{
		dir:    dir,
		filter: filter,
		branch: branch,
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
		if m.branchActive {
			return m.updateBranchPrompt(msg)
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "ctrl+p":
			return m.runOnFiltered(pullCmd)
		case "ctrl+o":
			return m.openBranchPrompt()
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
	case cmdDoneMsg:
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
func (m model) runOnFiltered(cmdFor func(name string, repo git.Repo) tea.Cmd) (model, tea.Cmd) {
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

// openBranchPrompt opens the transient checkout prompt. While it is open, key
// input edits the prompt (enter switches the filtered repos, esc cancels)
// rather than the filter.
func (m model) openBranchPrompt() (model, tea.Cmd) {
	m.branchActive = true
	m.branch.Reset()
	m.filter.Blur()
	return m, m.branch.Focus()
}

// updateBranchPrompt routes a key to the open checkout prompt.
func (m model) updateBranchPrompt(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.closeBranchPrompt()
	case "enter":
		branch := strings.TrimSpace(m.branch.Value())
		m, focusCmd := m.closeBranchPrompt()
		if branch == "" {
			return m, focusCmd
		}
		m, runCmd := m.runOnFiltered(checkoutCmd(branch))
		return m, tea.Batch(focusCmd, runCmd)
	}
	var cmd tea.Cmd
	m.branch, cmd = m.branch.Update(msg)
	return m, cmd
}

// closeBranchPrompt dismisses the checkout prompt and returns focus to the filter.
func (m model) closeBranchPrompt() (model, tea.Cmd) {
	m.branchActive = false
	m.branch.Blur()
	return m, m.filter.Focus()
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
	b.WriteString("\n")
	if m.branchActive {
		b.WriteString(m.branch.View())
		b.WriteString("\n")
	}
	b.WriteString("\n")

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
