package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/mattn/go-shellwords"
	"github.com/rs/zerolog/log"
)

type repoEntry struct {
	name   string
	repo   git.Repo
	status *repoStatus  // nil until loaded
	diff   *lineChanges // nil until loaded
	cmd    cmdState
	result *cmdResult // last command's full output; nil until one completes
}

// uiMode is which screen has key focus. In modeList the filter is focused and
// printable keys filter; modeCommand captures keys for the git command input;
// modeHelp shows the bindings overlay.
type uiMode int

const (
	modeList uiMode = iota
	modeCommand
	modeHelp
)

// model is the root TUI model. In the list, the filter input is always focused
// (fzf-style): printable keys edit the filter and every action is a
// non-printable binding. tab switches to command mode, where the same input
// line edits an arbitrary git command run against the filtered repos.
type model struct {
	dir    string
	filter textinput.Model
	// command is the git command input, shown and given key focus only in
	// modeCommand. The filter is blurred for its duration but keeps its value,
	// so the list stays filtered and the command targets that subset.
	command textinput.Model
	repos   []repoEntry
	cursor  int         // index into the filtered list
	mode    uiMode      // which screen owns key input
	field   filterField // which text the filter matches against
	width   int
	height  int
	// output is the scrollable pane showing the cursor repo's last command
	// output. It is rendered only when that repo's command failed; outputName
	// is the repo currently loaded into it, so scrolling persists until the
	// cursor moves to a different repo.
	output     viewport.Model
	outputName string
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
	command.Placeholder = "git command (runs on filtered repos)"
	command.SetWidth(40)
	return model{
		dir:     dir,
		filter:  filter,
		command: command,
		output:  viewport.New(),
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
		m.output.SetWidth(msg.Width)
		m.output.SetHeight(m.paneBodyHeight())
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
		case "up", "ctrl+k":
			m.cursor--
			return m.clampCursor().refreshOutput(), nil
		case "down", "ctrl+j":
			m.cursor++
			return m.clampCursor().refreshOutput(), nil
		case "pgup":
			m.output.PageUp()
			return m, nil
		case "pgdown":
			m.output.PageDown()
			return m, nil
		case "tab":
			return m.enterCommand()
		case "ctrl+1":
			m.field = fieldNameBranch
			return m.afterModeChange()
		case "ctrl+2":
			m.field = fieldName
			return m.afterModeChange()
		case "ctrl+3":
			m.field = fieldBranch
			return m.afterModeChange()
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
		return m.addRepo(msg.name, msg.repo), tea.Batch(statusCmd(msg.name, msg.repo), diffCmd(msg.name, msg.repo))
	case statusLoadedMsg:
		return m.setStatus(msg.name, msg.status), nil
	case diffLoadedMsg:
		return m.setDiff(msg.name, msg.changes), nil
	case cmdDoneMsg:
		m = m.setCmdResult(msg).refreshOutput()
		if repo, ok := m.repoByName(msg.name); ok {
			// auto-refresh status and line changes after the command
			return m, tea.Batch(statusCmd(msg.name, repo), diffCmd(msg.name, repo))
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	// the filter may have changed which rows match, and so which repo the pane targets
	m = m.clampCursor().refreshOutput()
	return m, cmd
}

// runOnFiltered marks every repo currently matching the filter as running and
// fires cmdFor against each. This is the shared entry point for command runs.
func (m model) runOnFiltered(cmdFor func(name string, repo git.Repo) tea.Cmd) (model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, i := range m.matchedIndexes() {
		m.repos[i].cmd = cmdRunning
		m.repos[i].result = nil // drop the prior output; the pane hides until this finishes
		cmds = append(cmds, cmdFor(m.repos[i].name, m.repos[i].repo))
	}
	return m.refreshOutput(), tea.Batch(cmds...)
}

// enterCommand switches to command mode. The filter is blurred but keeps its
// value, so the list stays filtered while the command input is focused.
func (m model) enterCommand() (model, tea.Cmd) {
	m.mode = modeCommand
	m.command.Reset()
	m.filter.Blur()
	return m, m.command.Focus()
}

// updateCommand routes a key while the command input is focused. enter parses
// the line (shell-style, "git" prefix optional), runs it on the filtered repos,
// then clears the line but stays in command mode so commands can be chained;
// tab is the sole switch back to the filter and esc quits. ↑/↓ and ctrl+k/j
// move the repo cursor and pgup/pgdn scroll the failure pane, so the cursor
// repo's output is inspectable without leaving command mode.
func (m model) updateCommand(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "tab":
		return m.exitCommand()
	case "up", "ctrl+k":
		m.cursor--
		return m.clampCursor().refreshOutput(), nil
	case "down", "ctrl+j":
		m.cursor++
		return m.clampCursor().refreshOutput(), nil
	case "pgup":
		m.output.PageUp()
		return m, nil
	case "pgdown":
		m.output.PageDown()
		return m, nil
	case "enter":
		input := strings.TrimSpace(m.command.Value())
		m.command.Reset()
		args, err := shellwords.Parse(input)
		if err != nil {
			log.Error().Err(err).Str("input", input).Msg("failed to parse command")
			return m, nil
		}
		if len(args) > 0 && args[0] == "git" {
			args = args[1:]
		}
		if len(args) == 0 {
			return m, nil
		}
		return m.runOnFiltered(func(name string, repo git.Repo) tea.Cmd {
			return commandCmd(name, repo, args)
		})
	}
	var cmd tea.Cmd
	m.command, cmd = m.command.Update(msg)
	return m, cmd
}

// exitCommand dismisses the command input and returns focus to the filter.
func (m model) exitCommand() (model, tea.Cmd) {
	m.mode = modeList
	m.command.Blur()
	return m, m.filter.Focus()
}

// afterModeChange applies a field switch: it refreshes the prompt to reflect the
// new mode, resizes the filter for the new prompt width, and clamps the cursor
// since the matched set may have changed.
func (m model) afterModeChange() (model, tea.Cmd) {
	m.filter.Prompt = m.filterPrompt()
	if m.width > 0 {
		m.filter.SetWidth(m.width - lipgloss.Width(m.filter.Prompt))
	}
	return m.clampCursor(), nil
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

// refreshFiltered re-fetches git status and line changes for every repo
// matching the filter.
func (m model) refreshFiltered() (model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, r := range m.matched() {
		cmds = append(cmds, statusCmd(r.name, r.repo), diffCmd(r.name, r.repo))
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

// setCmdResult records the outcome and full output of a command on the named
// repo. A non-nil err (non-zero exit or a failure to start) marks the row
// failed, which is what gates the output pane.
func (m model) setCmdResult(msg cmdDoneMsg) model {
	for i := range m.repos {
		if m.repos[i].name == msg.name {
			if msg.err != nil {
				m.repos[i].cmd = cmdFailed
			} else {
				m.repos[i].cmd = cmdOK
			}
			m.repos[i].result = &cmdResult{args: msg.args, exit: msg.exit, stdout: msg.stdout, stderr: msg.stderr}
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

// cursorOutput returns the repo under the cursor and its full command output,
// but only when that repo's last command failed. Otherwise it returns an empty
// name, which is the signal that no pane should show.
func (m model) cursorOutput() (string, string) {
	matched := m.matched()
	if m.cursor < 0 || m.cursor >= len(matched) {
		return "", ""
	}
	r := matched[m.cursor]
	if r.cmd != cmdFailed || r.result == nil {
		return "", ""
	}
	return r.name, r.result.body()
}

// refreshOutput points the output viewport at the cursor repo's failed-command
// output. It reloads (and resets the scroll to the top) only when the targeted
// repo changes, so scrolling within one repo's output survives re-renders.
func (m model) refreshOutput() model {
	name, content := m.cursorOutput()
	if name != m.outputName {
		m.outputName = name
		m.output.SetContent(content)
		m.output.GotoTop()
	}
	return m
}

// paneBodyHeight is the viewport height: the ~1/3-screen pane minus its
// separator and header lines.
func (m model) paneBodyHeight() int {
	return max(3, m.height/3-2)
}

// paneView renders the failure pane for the named repo: a separator, a header
// naming the repo and the command that ran, and the scrollable output.
func (m model) paneView(name string) string {
	r, ok := m.entryByName(name)
	if !ok || r.result == nil {
		return ""
	}
	header := fmt.Sprintf("%s $ git %s → exit %d", name, strings.Join(r.result.args, " "), r.result.exit)
	sep := strings.Repeat("─", max(1, m.width))
	return lipgloss.JoinVertical(lipgloss.Left, sep, header, m.output.View())
}

func (m model) entryByName(name string) (repoEntry, bool) {
	for i := range m.repos {
		if m.repos[i].name == name {
			return m.repos[i], true
		}
	}
	return repoEntry{}, false
}

func (m model) View() tea.View {
	if m.mode == modeHelp {
		return tea.View{Content: helpContent(), AltScreen: true}
	}

	input := m.filter.View()
	if m.mode == modeCommand {
		input = m.command.View()
	}

	top := lipgloss.JoinVertical(lipgloss.Left, input, "", m.listContent())
	name, _ := m.cursorOutput()
	if name == "" {
		return tea.View{Content: top, AltScreen: true}
	}

	// Anchor the failure pane to the bottom of the screen: pad the gap between
	// the list and the pane so the pane sits in the bottom ~third rather than
	// floating directly under a short list.
	pane := m.paneView(name)
	content := lipgloss.JoinVertical(lipgloss.Left, top, pane)
	if gap := m.height - lipgloss.Height(top) - lipgloss.Height(pane); gap > 0 {
		content = lipgloss.JoinVertical(lipgloss.Left, top, strings.Repeat("\n", gap-1), pane)
	}
	return tea.View{Content: content, AltScreen: true}
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

	nameWidth, statusWidth := 0, 0
	for _, r := range matched {
		if w := lipgloss.Width(r.name); w > nameWidth {
			nameWidth = w
		}
		if w := lipgloss.Width(statusText(r)); w > statusWidth {
			statusWidth = w
		}
	}
	nameCol := lipgloss.NewStyle().Width(nameWidth)
	statusCol := lipgloss.NewStyle().Width(statusWidth)

	rows := make([]string, len(matched))
	for i, r := range matched {
		marker := "  "
		if i == m.cursor {
			marker = "▸ "
		}
		changes := "..."
		if r.diff != nil {
			changes = r.diff.String()
		}
		cols := []string{marker, nameCol.Render(r.name), "  ", statusCol.Render(statusText(r)), "  ", changes}
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

// statusText is the status column for a row, or "..." until status loads.
func statusText(r repoEntry) string {
	if r.status == nil {
		return "..."
	}
	return r.status.line()
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
