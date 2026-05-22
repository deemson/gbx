package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/deemson/gbx/internal/git"
)

type repoEntry struct {
	name   string
	repo   git.Repo
	status *repoStatus // nil until loaded
	cmd    cmdState
	cmdErr error // last command error, surfaced via drill-in
}

// uiMode is which screen has key focus. In modeList the filter is focused and
// printable keys filter; the other modes capture keys for their own use.
type uiMode int

const (
	modeList uiMode = iota
	modeBranchPrompt
	modeDetail
	modeHelp
)

// model is the root TUI model. The filter input is always focused (fzf-style):
// printable keys edit the filter, and every action is a non-printable binding.
type model struct {
	dir    string
	filter textinput.Model
	// branch is the transient checkout prompt, shown and given key focus only
	// in modeBranchPrompt. The filter is blurred for its duration.
	branch textinput.Model
	repos  []repoEntry
	cursor int        // index into the filtered list; the drill-in target
	mode   uiMode     // which screen owns key input
	detail detailView // populated while mode == modeDetail
	width  int
	height int
}

func newModel(dir string) model {
	filter := textinput.New()
	filter.Prompt = "> "
	filter.Placeholder = "filter repos"
	// A non-zero width is required up front: textinput truncates the placeholder
	// to Width()+1 runes, so Width()==0 renders only its first letter. Resized to
	// the terminal on the first WindowSizeMsg.
	filter.SetWidth(40)
	// Focus here, in the constructor, so the focused state persists into the
	// model the program runs. Calling Focus() in Init() would not, because
	// Init() returns only a Cmd, discarding the mutated model.
	filter.Focus()
	branch := textinput.New()
	branch.Prompt = "branch: "
	branch.Placeholder = "switch to branch"
	branch.SetWidth(40)
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
		m.filter.SetWidth(msg.Width - lipgloss.Width(m.filter.Prompt))
		m.branch.SetWidth(msg.Width - lipgloss.Width(m.branch.Prompt))
		return m, nil
	case tea.KeyPressMsg:
		switch m.mode {
		case modeBranchPrompt:
			return m.updateBranchPrompt(msg)
		case modeDetail:
			return m.updateDetail(msg)
		case modeHelp:
			return m.updateHelp(msg)
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "ctrl+k":
			m.cursor--
			return m.clampCursor(), nil
		case "down", "ctrl+j":
			m.cursor++
			return m.clampCursor(), nil
		case "enter":
			return m.openDetail()
		case "ctrl+p":
			return m.runOnFiltered(pullCmd)
		case "ctrl+o":
			return m.openBranchPrompt()
		case "ctrl+r":
			return m.refreshFiltered()
		case "ctrl+g":
			m.mode = modeHelp
			return m, nil
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
	case detailLoadedMsg:
		if m.mode == modeDetail && m.detail.name == msg.name {
			m.detail.diff = msg.diff
			m.detail.err = msg.err
			m.detail.loaded = true
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m = m.clampCursor() // the filter may have changed which rows match
	return m, cmd
}

// runOnFiltered marks every repo currently matching the filter as running and
// fires cmdFor against each. This is the shared entry point for command
// bindings.
func (m model) runOnFiltered(cmdFor func(name string, repo git.Repo) tea.Cmd) (model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, i := range m.matchedIndexes() {
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
	m.mode = modeBranchPrompt
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
	m.mode = modeList
	m.branch.Blur()
	return m, m.filter.Focus()
}

// matchedIndexes returns indexes into m.repos passing the filter, ranked
// best-match-first; an empty filter yields every index in display order.
func (m model) matchedIndexes() []int {
	names := make([]string, len(m.repos))
	for i, r := range m.repos {
		names[i] = r.name
	}
	return rankFilter(m.filter.Value(), names)
}

// matched returns the repos currently passing the filter, ranked best-match-first.
func (m model) matched() []repoEntry {
	idx := m.matchedIndexes()
	out := make([]repoEntry, len(idx))
	for i, j := range idx {
		out[i] = m.repos[j]
	}
	return out
}

// clampCursor keeps the cursor within the filtered list as it grows or shrinks.
func (m model) clampCursor() model {
	switch n := len(m.matched()); {
	case n == 0, m.cursor < 0:
		m.cursor = 0
	case m.cursor >= n:
		m.cursor = n - 1
	}
	return m
}

// openDetail drills into the repo under the cursor, loading its diff vs HEAD.
func (m model) openDetail() (model, tea.Cmd) {
	m = m.clampCursor()
	matched := m.matched()
	if len(matched) == 0 {
		return m, nil
	}
	sel := matched[m.cursor]
	m.mode = modeDetail
	m.detail = detailView{name: sel.name, cmdErr: sel.cmdErr}
	return m, detailCmd(sel.name, sel.repo)
}

// updateDetail routes a key while the drill-in is open. Esc returns to the list.
func (m model) updateDetail(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = modeList
		return m, nil
	}
	return m, nil
}

// refreshFiltered re-fetches git status for every repo matching the filter.
func (m model) refreshFiltered() (model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, r := range m.matched() {
		cmds = append(cmds, statusCmd(r.name, r.repo))
	}
	return m, tea.Batch(cmds...)
}

// updateHelp routes a key while the help overlay is open. Esc or ctrl+g closes it.
func (m model) updateHelp(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "ctrl+g":
		m.mode = modeList
		return m, nil
	}
	return m, nil
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
	if m.mode == modeDetail {
		return tea.View{Content: m.detailContent(), AltScreen: true}
	}
	if m.mode == modeHelp {
		return tea.View{Content: helpContent(), AltScreen: true}
	}

	sections := []string{m.filter.View()}
	if m.mode == modeBranchPrompt {
		sections = append(sections, m.branch.View())
	}
	sections = append(sections, "", m.listContent())

	return tea.View{
		Content:   lipgloss.JoinVertical(lipgloss.Left, sections...),
		AltScreen: true,
	}
}

// listContent renders the repo rows (or an empty-state line) as a single block.
func (m model) listContent() string {
	if len(m.repos) == 0 {
		return "no repos"
	}
	matched := m.matched()
	if len(matched) == 0 {
		return "no matches"
	}

	nameWidth := 0
	for _, r := range matched {
		if w := lipgloss.Width(r.name); w > nameWidth {
			nameWidth = w
		}
	}
	nameCol := lipgloss.NewStyle().Width(nameWidth)

	rows := make([]string, len(matched))
	for i, r := range matched {
		marker := "  "
		if i == m.cursor {
			marker = "▸ "
		}
		status := "..."
		if r.status != nil {
			status = r.status.line()
		}
		cols := []string{marker, nameCol.Render(r.name), "  ", status}
		if g := r.cmd.glyph(); g != "" {
			cols = append(cols, "  ", g)
		}
		rows[i] = lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// detailContent renders the drill-in screen for m.detail.
func (m model) detailContent() string {
	d := m.detail
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", d.name)

	b.WriteString("changes vs HEAD:\n")
	switch {
	case !d.loaded:
		b.WriteString("  ...\n")
	case d.err != nil:
		fmt.Fprintf(&b, "  (error: %s)\n", d.err)
	case len(d.diff.Paths) == 0:
		b.WriteString("  none\n")
	default:
		for _, p := range d.diff.Paths {
			fmt.Fprintf(&b, "  +%d -%d  %s\n", p.AddedLines, p.DeletedLines, p.Path)
		}
	}

	if d.cmdErr != nil {
		fmt.Fprintf(&b, "\nlast command error:\n  %s\n", d.cmdErr)
	}

	b.WriteString("\nesc: back\n")
	return b.String()
}
