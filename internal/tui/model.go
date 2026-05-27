package tui

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/deemson/gbx/internal/git"
)

type repoEntry struct {
	name     string
	repo     git.Repo
	status   *repoStatus  // nil until loaded
	diff     *lineChanges // nil until loaded
	branches []string     // nil until loaded; feeds checkout autocomplete
	cmd      cmdState
	cmdErr   error // last command's error; nil on success. Drives the row one-liner.
}

// uiMode is which screen has key focus. modeFilter focuses the filter (printable
// keys filter, fzf-style); modeCommand focuses the command line with its
// autocomplete; modeHelp shows the bindings overlay.
type uiMode int

const (
	modeFilter uiMode = iota
	modeCommand
	modeHelp
)

// model is the root TUI model. It starts in filter mode (the filter input is
// focused, printable keys filter). enter applies the filter and switches to
// command mode, where the same narrowed list is acted on by one of four typed
// git commands typed into the command line with position-aware autocomplete.
type model struct {
	dir    string
	filter textinput.Model
	// command is the git command input, focused only in modeCommand. The filter
	// is blurred but keeps its value, so the list stays narrowed and the command
	// targets that subset.
	command textinput.Model
	repos   []repoEntry
	mode    uiMode
	field   filterField // which text the filter matches against
	width   int
	height  int
	// suggestions are the autocomplete options for the command line's active
	// token; suggIndex is the highlighted one, or -1 when none is applied yet.
	suggestions []string
	suggIndex   int
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
	command := textinput.New()
	command.Prompt = "$ "
	command.Placeholder = "checkout <ref> · checkout -b <name> · fetch · pull"
	command.SetWidth(40)
	return model{
		dir:       dir,
		filter:    filter,
		command:   command,
		suggIndex: -1,
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
		m.command.SetWidth(msg.Width - lipgloss.Width(m.command.Prompt))
		return m, nil
	case tea.KeyPressMsg:
		switch m.mode {
		case modeCommand:
			return m.updateCommand(msg)
		case modeHelp:
			return m.updateHelp(msg)
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.enterCommand()
		case "ctrl+1":
			m.field = fieldNameBranch
			return m.afterFieldChange()
		case "ctrl+2":
			m.field = fieldName
			return m.afterFieldChange()
		case "ctrl+3":
			m.field = fieldBranch
			return m.afterFieldChange()
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
		return m.addRepo(msg.name, msg.repo), tea.Batch(
			statusCmd(msg.name, msg.repo), diffCmd(msg.name, msg.repo), branchesCmd(msg.name, msg.repo))
	case statusLoadedMsg:
		return m.setStatus(msg.name, msg.status), nil
	case diffLoadedMsg:
		return m.setDiff(msg.name, msg.changes), nil
	case branchesLoadedMsg:
		return m.setBranches(msg.name, msg.branches), nil
	case cmdDoneMsg:
		m = m.setCmdDone(msg)
		if repo, ok := m.repoByName(msg.name); ok {
			// auto-refresh status, line changes, and branches after the command
			return m, tea.Batch(statusCmd(msg.name, repo), diffCmd(msg.name, repo), branchesCmd(msg.name, repo))
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	return m, cmd
}

// runOnFiltered marks every repo currently matching the filter as running and
// fires cmdFor against each. This is the shared entry point for command runs.
func (m model) runOnFiltered(cmdFor func(name string, repo git.Repo) tea.Cmd) (model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, i := range m.matchedIndexes() {
		m.repos[i].cmd = cmdRunning
		m.repos[i].cmdErr = nil
		cmds = append(cmds, cmdFor(m.repos[i].name, m.repos[i].repo))
	}
	return m, tea.Batch(cmds...)
}

// enterCommand applies the current filter and switches to command mode. The
// filter is blurred but keeps its value, so the list stays narrowed while the
// command input is focused.
func (m model) enterCommand() (model, tea.Cmd) {
	m.mode = modeCommand
	m.command.Reset()
	m = m.recomputeSuggestions()
	m.filter.Blur()
	return m, m.command.Focus()
}

// updateCommand routes a key while the command input is focused. enter runs the
// parsed command on the filtered repos and clears the line, staying in command
// mode; tab / shift+tab cycle the autocomplete suggestions, writing the
// highlighted one into the line; esc returns to filter mode and clears the
// filter; ctrl+c quits.
func (m model) updateCommand(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.exitToFilter()
	case "tab":
		return m.cycleSuggestion(1), nil
	case "shift+tab":
		return m.cycleSuggestion(-1), nil
	case "enter":
		return m.submitCommand()
	}
	var cmd tea.Cmd
	m.command, cmd = m.command.Update(msg)
	// Editing the line changes the active token, so recompute the suggestions.
	m = m.recomputeSuggestions()
	return m, cmd
}

// submitCommand parses the command line and, when it names one of the four
// supported commands, runs it on the filtered repos. The line is cleared and we
// stay in command mode either way; an unrecognized line is a no-op.
func (m model) submitCommand() (model, tea.Cmd) {
	fields := strings.Fields(m.command.Value())
	m.command.Reset()
	m = m.recomputeSuggestions()
	action, ok := parseCommand(fields)
	if !ok {
		return m, nil
	}
	return m.runOnFiltered(action)
}

// recomputeSuggestions refreshes the suggestion set for the command line's
// current active token and resets the highlight (the next tab starts at the
// first suggestion).
func (m model) recomputeSuggestions() model {
	m.suggestions = m.suggestionsFor(m.command.Value())
	m.suggIndex = -1
	return m
}

// cycleSuggestion advances the highlighted suggestion by delta (with wrap) and
// writes it into the command line, replacing the active token. The suggestion
// set itself is left untouched so repeated tabs keep cycling the same options.
func (m model) cycleSuggestion(delta int) model {
	n := len(m.suggestions)
	if n == 0 {
		return m
	}
	switch {
	case delta > 0:
		m.suggIndex = (m.suggIndex + 1) % n
	case m.suggIndex <= 0:
		m.suggIndex = n - 1
	default:
		m.suggIndex--
	}
	head, _ := splitActive(m.command.Value())
	m.command.SetValue(head + m.suggestions[m.suggIndex])
	m.command.CursorEnd()
	return m
}

// exitToFilter dismisses the command input, clears the filter, and returns focus
// to it.
func (m model) exitToFilter() (model, tea.Cmd) {
	m.mode = modeFilter
	m.command.Blur()
	m.filter.Reset()
	return m, m.filter.Focus()
}

// afterFieldChange applies a field switch: it refreshes the prompt to reflect
// the new field and resizes the filter for the new prompt width.
func (m model) afterFieldChange() (model, tea.Cmd) {
	m.filter.Prompt = m.filterPrompt()
	if m.width > 0 {
		m.filter.SetWidth(m.width - lipgloss.Width(m.filter.Prompt))
	}
	return m, nil
}

// filterPrompt encodes the active field mode: the field name prefixes the ">"
// glyph, except for the default name+branch.
func (m model) filterPrompt() string {
	switch m.field {
	case fieldName:
		return "name > "
	case fieldBranch:
		return "branch > "
	default:
		return "> "
	}
}

// matchedIndexes returns indexes into m.repos passing the filter, ranked
// best-match-first; an empty filter yields every index in display order. The
// active field (ctrl+1..3) and the fzf-style filter DSL shape the match. A repo
// whose status has not loaded contributes an empty branch, which never matches.
func (m model) matchedIndexes() []int {
	names := make([]string, len(m.repos))
	branches := make([]string, len(m.repos))
	for i, r := range m.repos {
		names[i] = r.name
		if r.status != nil {
			branches[i] = r.status.branch
		}
	}
	return rankFilter(m.filter.Value(), names, branches, m.field)
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

// refreshFiltered re-fetches git status, line changes, and branches for every
// repo matching the filter.
func (m model) refreshFiltered() (model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, r := range m.matched() {
		cmds = append(cmds, statusCmd(r.name, r.repo), diffCmd(r.name, r.repo), branchesCmd(r.name, r.repo))
	}
	return m, tea.Batch(cmds...)
}

// updateHelp routes a key while the help overlay is open. Esc or ctrl+g closes it.
func (m model) updateHelp(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "ctrl+g":
		m.mode = modeFilter
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

// setDiff attaches loaded line changes to the named repo.
func (m model) setDiff(name string, changes lineChanges) model {
	for i := range m.repos {
		if m.repos[i].name == name {
			loaded := changes
			m.repos[i].diff = &loaded
			break
		}
	}
	return m
}

// setBranches attaches a loaded branch list to the named repo.
func (m model) setBranches(name string, branches []string) model {
	for i := range m.repos {
		if m.repos[i].name == name {
			m.repos[i].branches = branches
			break
		}
	}
	return m
}

// setCmdDone records a command's outcome on the named repo: a non-nil err marks
// the row failed (which surfaces the error as the row one-liner), nil marks it OK.
func (m model) setCmdDone(msg cmdDoneMsg) model {
	for i := range m.repos {
		if m.repos[i].name == msg.name {
			if msg.err != nil {
				m.repos[i].cmd = cmdFailed
				m.repos[i].cmdErr = msg.err
			} else {
				m.repos[i].cmd = cmdOK
				m.repos[i].cmdErr = nil
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
	if m.mode == modeHelp {
		return tea.View{Content: helpContent(), AltScreen: true}
	}
	var content string
	if m.mode == modeCommand {
		content = lipgloss.JoinVertical(lipgloss.Left, m.command.View(), m.suggestionLine(), "", m.listContent())
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), "", m.listContent())
	}
	return tea.View{Content: content, AltScreen: true}
}

// suggestionLine renders the autocomplete options beneath the command line, the
// highlighted one (cycled by tab) reversed. Empty when there is nothing to
// suggest for the active token.
func (m model) suggestionLine() string {
	if len(m.suggestions) == 0 {
		return ""
	}
	selected := lipgloss.NewStyle().Reverse(true)
	parts := make([]string, len(m.suggestions))
	for i, s := range m.suggestions {
		if i == m.suggIndex {
			parts[i] = selected.Render(s)
		} else {
			parts[i] = colorDim.Render(s)
		}
	}
	line := strings.Join(parts, "  ")
	if m.width > 0 {
		line = ansi.Truncate(line, m.width, "…")
	}
	return line
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

	// Column widths are pinned to the full repo list, not just the matched
	// subset, so they stay put as the filter narrows the visible rows.
	nameWidth, branchWidth, stateWidth := 0, 0, 0
	for _, r := range m.repos {
		if w := lipgloss.Width(r.name); w > nameWidth {
			nameWidth = w
		}
		if w := lipgloss.Width(branchText(r)); w > branchWidth {
			branchWidth = w
		}
		if w := lipgloss.Width(stateText(r)); w > stateWidth {
			stateWidth = w
		}
	}
	nameCol := lipgloss.NewStyle().Width(nameWidth)
	branchCol := lipgloss.NewStyle().Width(branchWidth)
	stateCol := lipgloss.NewStyle().Width(stateWidth)

	rows := make([]string, len(matched))
	for i, r := range matched {
		cols := []string{nameCol.Render(r.name), "  ", branchCol.Render(branchText(r)), "  ", stateCol.Render(stateText(r))}
		switch {
		case r.diff == nil:
			cols = append(cols, "  ", "...") // line changes not loaded yet
		case !r.diff.empty():
			cols = append(cols, "  ", r.diff.String()) // hidden entirely when +0 -0
		}
		if g := r.cmd.glyph(); g != "" {
			cols = append(cols, "  ", g)
		}
		if s := r.summary(); s != "" {
			prefix := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
			if m.width > 0 {
				s = truncate(s, m.width-lipgloss.Width(prefix)-2)
			}
			cols = append(cols, "  ", s)
		}
		rows[i] = lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// branchText is the branch column for a row, or "..." until status loads.
func branchText(r repoEntry) string {
	if r.status == nil {
		return "..."
	}
	return r.status.branchField()
}

// stateText is the change-state column for a row, empty until status loads.
func stateText(r repoEntry) string {
	if r.status == nil {
		return ""
	}
	return r.status.stateField()
}

// truncate clips s to at most max runes, ending in "…" when it had to cut.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}
