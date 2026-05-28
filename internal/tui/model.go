package tui

import (
	"context"
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

// uiMode is which screen has key focus. modeList is the default — letter keys
// trigger commands directly; F1/F4/c/b open transient overlays. The three
// *Prompt modes own the bottom-row text input; modeHelp is the alt-screen
// bindings overlay.
type uiMode int

const (
	modeList uiMode = iota
	modeFilterPrompt
	modeCheckoutPrompt
	modeBranchPrompt
	modeHelp
)

// model is the root TUI model. List mode is the default state; letter keys
// (r/f/p/c/b/q) dispatch directly to git actions on the filtered set, F4 opens a
// transient filter prompt (committed → m.filter on Enter; reverted on F4 or
// ESC-on-empty), c/b open argument prompts. The prompt textinput is shared
// across the three prompt modes — its label and Enter semantics vary by mode.
type model struct {
	dir    string
	repos  []repoEntry
	mode   uiMode
	field  filterField
	width  int
	height int

	// filter is the committed filter pattern applied to the visible row set
	// while in list mode (and while c/b prompts are open). The filter prompt's
	// draft (prompt.Value()) takes over live while modeFilterPrompt is active,
	// so what you'd commit is what you see.
	filter string

	// prompt is the shared bottom-row input, focused only in *Prompt modes.
	prompt textinput.Model

	// suggestions / suggIndex back the checkout-prompt branch autocomplete.
	suggestions []string
	suggIndex   int
}

func newModel(dir string) model {
	p := textinput.New()
	p.Prompt = "filter: "
	// A non-zero width is required up front: textinput truncates the placeholder
	// to Width()+1 runes. Resized on the first WindowSizeMsg.
	p.SetWidth(40)
	return model{
		dir:       dir,
		prompt:    p,
		suggIndex: -1,
	}
}

func (m model) Init() tea.Cmd {
	return readEntriesCmd(m.dir)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.prompt.SetWidth(msg.Width - lipgloss.Width(m.prompt.Prompt))
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(msg)
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
		m = m.setBranches(msg.name, msg.branches)
		if m.mode == modeCheckoutPrompt {
			m = m.recomputeSuggestions()
		}
		return m, nil
	case cmdDoneMsg:
		m = m.setCmdDone(msg)
		if repo, ok := m.repoByName(msg.name); ok {
			return m, tea.Batch(statusCmd(msg.name, repo), diffCmd(msg.name, repo), branchesCmd(msg.name, repo))
		}
		return m, nil
	}
	return m, nil
}

func (m model) updateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	switch m.mode {
	case modeFilterPrompt:
		nm, cmd := m.updateFilterPrompt(msg)
		return nm, cmd
	case modeCheckoutPrompt:
		nm, cmd := m.updateCheckoutPrompt(msg)
		return nm, cmd
	case modeBranchPrompt:
		nm, cmd := m.updateBranchPrompt(msg)
		return nm, cmd
	case modeHelp:
		return m.updateHelp(msg), nil
	}
	return m.updateList(msg)
}

// updateList routes a key in list mode (default). Letter keys dispatch git
// actions on the filtered set; F1/F4 open overlays; c/b open argument prompts;
// q quits; ctrl+1/2/3 toggle the filter field.
func (m model) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f1":
		m.mode = modeHelp
		return m, nil
	case "f4":
		return m.openFilterPrompt()
	case "q":
		return m, tea.Quit
	case "r":
		return m.refreshFiltered()
	case "f":
		return m.runOnFiltered(func(name string, repo git.Repo) tea.Cmd {
			return runCmd(name, "fetch", repo.Fetch)
		})
	case "p":
		return m.runOnFiltered(func(name string, repo git.Repo) tea.Cmd {
			return runCmd(name, "pull", repo.PullFastForward)
		})
	case "c":
		return m.openCheckoutPrompt()
	case "b":
		return m.openBranchPrompt()
	case "ctrl+1":
		m.field = fieldNameBranch
		return m, nil
	case "ctrl+2":
		m.field = fieldName
		return m, nil
	case "ctrl+3":
		m.field = fieldBranch
		return m, nil
	}
	return m, nil
}

// openFilterPrompt enters the F4 prompt with the committed filter pre-filled,
// so editing starts from the current state. effectiveFilter() makes the visible
// rows track the draft live while the prompt is open.
func (m model) openFilterPrompt() (model, tea.Cmd) {
	m.mode = modeFilterPrompt
	m = m.applyPromptLabel(m.filterLabel())
	m.prompt.SetValue(m.filter)
	m.prompt.CursorEnd()
	return m, m.prompt.Focus()
}

// filterLabel reflects the active field in the filter prompt's label.
func (m model) filterLabel() string {
	switch m.field {
	case fieldName:
		return "name: "
	case fieldBranch:
		return "branch: "
	default:
		return "filter: "
	}
}

// updateFilterPrompt handles keys with the filter prompt focused. Enter commits
// the draft to m.filter and closes; F4 reverts (discards the draft) and closes;
// ESC clears the draft, or — when the draft is already empty — reverts and
// closes (same as F4-while-open). ctrl+1/2/3 still toggle the field; the
// prompt's label updates so it's visible.
func (m model) updateFilterPrompt(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "f4":
		return m.closePrompt(), nil
	case "enter":
		m.filter = m.prompt.Value()
		return m.closePrompt(), nil
	case "esc":
		if m.prompt.Value() == "" {
			return m.closePrompt(), nil
		}
		m.prompt.SetValue("")
		return m, nil
	case "ctrl+1":
		m.field = fieldNameBranch
		return m.applyPromptLabel(m.filterLabel()), nil
	case "ctrl+2":
		m.field = fieldName
		return m.applyPromptLabel(m.filterLabel()), nil
	case "ctrl+3":
		m.field = fieldBranch
		return m.applyPromptLabel(m.filterLabel()), nil
	}
	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

// openCheckoutPrompt enters the c prompt: empty draft, branch suggestions
// populated from the visible repos.
func (m model) openCheckoutPrompt() (model, tea.Cmd) {
	m.mode = modeCheckoutPrompt
	m = m.applyPromptLabel("checkout: ")
	m.prompt.SetValue("")
	m = m.recomputeSuggestions()
	return m, m.prompt.Focus()
}

// updateCheckoutPrompt handles keys with the c prompt focused. Enter runs
// `checkout <ref>` on the filtered repos and closes; ESC clears the draft or —
// when empty — reverts and closes. Tab/shift+tab cycle branch suggestions
// inline. There's no retrigger-close: `c` is a letter you'll want to type in a
// ref (e.g. `claude-branch`), unlike the non-typeable F4.
func (m model) updateCheckoutPrompt(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		return m.cycleSuggestion(1), nil
	case "shift+tab":
		return m.cycleSuggestion(-1), nil
	case "enter":
		ref := strings.TrimSpace(m.prompt.Value())
		if ref == "" {
			return m, nil
		}
		nm, cmd := m.runOnFiltered(func(name string, repo git.Repo) tea.Cmd {
			return runCmd(name, "checkout", func(ctx context.Context) error { return repo.Checkout(ctx, ref) })
		})
		return nm.(model).closePrompt(), cmd
	case "esc":
		if m.prompt.Value() == "" {
			return m.closePrompt(), nil
		}
		m.prompt.SetValue("")
		m = m.recomputeSuggestions()
		return m, nil
	}
	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	m = m.recomputeSuggestions()
	return m, cmd
}

// openBranchPrompt enters the b prompt: empty draft, no autocomplete (you're
// inventing a name).
func (m model) openBranchPrompt() (model, tea.Cmd) {
	m.mode = modeBranchPrompt
	m = m.applyPromptLabel("checkout -b: ")
	m.prompt.SetValue("")
	return m, m.prompt.Focus()
}

// updateBranchPrompt handles keys with the b prompt focused. Enter runs
// `checkout -b <name>` on the filtered repos and closes; ESC clears the draft
// or — when empty — reverts and closes. No retrigger-close (same reason as the
// c prompt: `b` is a typeable letter).
func (m model) updateBranchPrompt(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.prompt.Value())
		if name == "" {
			return m, nil
		}
		nm, cmd := m.runOnFiltered(func(repoName string, repo git.Repo) tea.Cmd {
			return runCmd(repoName, "checkout -b", func(ctx context.Context) error { return repo.CheckoutBranch(ctx, name) })
		})
		return nm.(model).closePrompt(), cmd
	case "esc":
		if m.prompt.Value() == "" {
			return m.closePrompt(), nil
		}
		m.prompt.SetValue("")
		return m, nil
	}
	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

// closePrompt blurs the prompt and returns to list mode, discarding any draft
// and clearing suggestions. Callers that commit a draft (Enter on the filter
// prompt) write to m.filter first.
func (m model) closePrompt() model {
	m.mode = modeList
	m.prompt.Blur()
	m.prompt.SetValue("")
	m.suggestions = nil
	m.suggIndex = -1
	return m
}

// applyPromptLabel sets the prompt's label and resizes the input to fill the
// row given the new label width.
func (m model) applyPromptLabel(label string) model {
	m.prompt.Prompt = label
	if m.width > 0 {
		m.prompt.SetWidth(m.width - lipgloss.Width(m.prompt.Prompt))
	}
	return m
}

// updateHelp handles keys with the help overlay open. F1 or ESC closes it.
func (m model) updateHelp(msg tea.KeyPressMsg) model {
	switch msg.String() {
	case "f1", "esc":
		m.mode = modeList
	}
	return m
}

// effectiveFilter is the filter string used by matchedIndexes — the prompt's
// live draft while the filter prompt is open, the committed value otherwise.
func (m model) effectiveFilter() string {
	if m.mode == modeFilterPrompt {
		return m.prompt.Value()
	}
	return m.filter
}

// runOnFiltered marks every repo currently matching the filter as running and
// fires cmdFor against each.
func (m model) runOnFiltered(cmdFor func(name string, repo git.Repo) tea.Cmd) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, i := range m.matchedIndexes() {
		m.repos[i].cmd = cmdRunning
		m.repos[i].cmdErr = nil
		cmds = append(cmds, cmdFor(m.repos[i].name, m.repos[i].repo))
	}
	return m, tea.Batch(cmds...)
}

// matchedIndexes returns the repo indexes passing the filter, ranked best-first.
// effectiveFilter() picks draft vs committed; a repo whose status has not loaded
// contributes an empty branch, which never matches.
func (m model) matchedIndexes() []int {
	names := make([]string, len(m.repos))
	branches := make([]string, len(m.repos))
	for i, r := range m.repos {
		names[i] = r.name
		if r.status != nil {
			branches[i] = r.status.branch
		}
	}
	return rankFilter(m.effectiveFilter(), names, branches, m.field)
}

// matched returns the repos currently passing the filter, ranked best-first.
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
func (m model) refreshFiltered() (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, r := range m.matched() {
		cmds = append(cmds, statusCmd(r.name, r.repo), diffCmd(r.name, r.repo), branchesCmd(r.name, r.repo))
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

// visibleBranches returns every distinct branch across the repos passing the
// filter, deduped and sorted — the source list for c-prompt autocomplete.
func (m model) visibleBranches() []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range m.matched() {
		for _, b := range r.branches {
			if !seen[b] {
				seen[b] = true
				out = append(out, b)
			}
		}
	}
	sort.Strings(out)
	return out
}

// recomputeSuggestions refreshes the c-prompt's suggestion set against the
// current draft and resets the highlight (tab starts at the first suggestion).
func (m model) recomputeSuggestions() model {
	m.suggestions = m.filteredBranches(m.prompt.Value())
	m.suggIndex = -1
	return m
}

// filteredBranches returns visibleBranches filtered by fuzzy match against the
// active token. An empty token returns the full list.
func (m model) filteredBranches(active string) []string {
	cands := m.visibleBranches()
	if active == "" {
		return cands
	}
	return fuzzyPick(active, cands)
}

// cycleSuggestion advances the highlighted suggestion by delta (with wrap) and
// writes it into the prompt.
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
	m.prompt.SetValue(m.suggestions[m.suggIndex])
	m.prompt.CursorEnd()
	return m
}

func (m model) View() tea.View {
	if m.mode == modeHelp {
		return tea.View{Content: helpContent(), AltScreen: true}
	}
	list := m.listContent()
	bar := m.bottomBar()

	// The branch suggestion line, if any, sits on the row above the bar.
	var middle []string
	if m.mode == modeCheckoutPrompt && len(m.suggestions) > 0 {
		middle = append(middle, m.suggestionLine())
	}

	if m.height > 0 {
		listHeight := m.height - 1 - len(middle)
		if listHeight < 1 {
			listHeight = 1
		}
		listArea := lipgloss.NewStyle().Height(listHeight).Render(list)
		parts := append([]string{listArea}, middle...)
		parts = append(parts, bar)
		return tea.View{Content: lipgloss.JoinVertical(lipgloss.Left, parts...), AltScreen: true}
	}
	parts := []string{list}
	parts = append(parts, middle...)
	parts = append(parts, bar)
	return tea.View{Content: lipgloss.JoinVertical(lipgloss.Left, parts...), AltScreen: true}
}

// bottomBar is the always-visible row at the bottom: the active prompt or, when
// no prompt is open, the committed filter (or empty); "F1 Help" pinned right.
func (m model) bottomBar() string {
	left := m.barLeft()
	right := colorDim.Render("F1 Help")
	if m.width <= 0 {
		return left + "  " + right
	}
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// barLeft is the left portion of the bottom bar: the prompt input while a
// prompt is open, the committed filter status when not.
func (m model) barLeft() string {
	switch m.mode {
	case modeFilterPrompt, modeCheckoutPrompt, modeBranchPrompt:
		return m.prompt.View()
	}
	if m.filter == "" {
		return ""
	}
	return colorDim.Render(m.filterLabel() + m.filter)
}

// suggestionLine renders the c-prompt's autocomplete options, the highlighted
// one (cycled by tab) reversed.
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
