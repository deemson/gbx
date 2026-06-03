package tui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
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

	// loading counts the in-flight status/diff/branches reads for this repo;
	// >0 means the row is busy reading and spins. loadErr is the last *load
	// cycle's* settled error: reset to nil when a cycle is dispatched, written
	// by any failing read, read once loading hits 0.
	loading int
	loadErr error
}

// loadFailedMsg signals that one of a repo's status/diff/branches reads failed.
// It decrements the in-flight counter (like the *LoadedMsg success path) and
// records the error as the cycle's loadErr, so the row can settle to ✗.
type loadFailedMsg struct {
	name string
	err  error
}

// uiMode is which screen has key focus. modeList is the default — letter keys
// trigger commands directly; ?/ctrl+f/c/b open transient overlays. The three
// *Prompt modes own the header's top-row text input; modeHelp is the alt-screen
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
// (r/f/p/c/b/q) dispatch directly to git actions on the filtered set, ctrl+f
// opens a transient filter prompt (committed → m.filter on Enter; reverted on
// ctrl+f or ESC-on-empty), c/b open argument prompts. The prompt textinput is
// shared across the three prompt modes — its label and Enter semantics vary by
// mode.
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

	// version / pid are static header chrome shown dim in the right corner.
	// version defaults to "dev"; release builds override it via WithVersion.
	version string
	pid     int

	// logPath is the per-PID log file, shown in the help overlay's header so a
	// failed session can be found. Set from WithLogPath (main.go owns the path).
	logPath string

	// help is the scrollable viewport for the ? overlay. Its content is static
	// (the bindings), set once; it's resized on every WindowSizeMsg and reset to
	// the top each time help opens.
	help viewport.Model

	// spinner is the single shared loading spinner rendered in every busy row's
	// left gutter. spinning guards its tick loop: the tick is kicked once when
	// work starts and stops itself when nothing is busy (see kickSpinner / the
	// spinner.TickMsg handler), so an idle app doesn't redraw 12×/second.
	spinner  spinner.Model
	spinning bool
}

func newModel(dir string) model {
	p := textinput.New()
	p.Prompt = filterLabel
	// A non-zero width is required up front: textinput truncates the placeholder
	// to Width()+1 runes. Resized on the first WindowSizeMsg.
	p.SetWidth(40)
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = colorDim
	hv := viewport.New()
	hv.SetContent(helpContent())
	return model{
		dir:       dir,
		prompt:    p,
		suggIndex: -1,
		version:   "dev",
		pid:       os.Getpid(),
		spinner:   sp,
		help:      hv,
	}
}

// Prompt labels in the header's top row. The label visually anchors what the
// row currently is — filter status, branch picker, or new-branch namer.
const (
	filterLabel       = "Filter: "
	switchBranchLabel = "Switch Branch: "
	newBranchLabel    = "New Branch: "
)

func (m model) Init() tea.Cmd {
	return readEntriesCmd(m.dir)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.prompt.SetWidth(msg.Width - lipgloss.Width(m.prompt.Prompt))
		m.help.SetWidth(msg.Width)
		m.help.SetHeight(msg.Height - lipgloss.Height(m.helpHeader()) - lipgloss.Height(m.helpFooter()))
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	case tea.MouseWheelMsg:
		if m.mode == modeHelp {
			var cmd tea.Cmd
			m.help, cmd = m.help.Update(msg)
			return m, cmd
		}
		return m, nil
	case entriesLoadedMsg:
		cmds := make([]tea.Cmd, 0, len(msg.entries))
		for _, e := range msg.entries {
			cmds = append(cmds, openRepoCmd(m.dir, e))
		}
		return m, tea.Batch(cmds...)
	case repoFoundMsg:
		m = m.addRepo(msg.name, msg.repo).startLoad(msg.name)
		var tick tea.Cmd
		m, tick = m.kickSpinner()
		return m, tea.Batch(
			statusCmd(msg.name, msg.repo), diffCmd(msg.name, msg.repo), branchesCmd(msg.name, msg.repo), tick)
	case statusLoadedMsg:
		return m.setStatus(msg.name, msg.status).loadDone(msg.name, nil), nil
	case diffLoadedMsg:
		return m.setDiff(msg.name, msg.changes).loadDone(msg.name, nil), nil
	case branchesLoadedMsg:
		m = m.setBranches(msg.name, msg.branches).loadDone(msg.name, nil)
		if m.mode == modeCheckoutPrompt || m.mode == modeBranchPrompt {
			m = m.recomputeSuggestions()
		}
		return m, nil
	case loadFailedMsg:
		return m.loadDone(msg.name, msg.err), nil
	case spinner.TickMsg:
		if !m.anyBusy() {
			m.spinning = false
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case cmdDoneMsg:
		m = m.setCmdDone(msg)
		if repo, ok := m.repoByName(msg.name); ok {
			m = m.startLoad(msg.name)
			var tick tea.Cmd
			m, tick = m.kickSpinner()
			return m, tea.Batch(statusCmd(msg.name, repo), diffCmd(msg.name, repo), branchesCmd(msg.name, repo), tick)
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
		return m.updateHelp(msg)
	}
	return m.updateList(msg)
}

// updateList routes a key in list mode (default). Letter keys dispatch git
// actions on the filtered set; ? toggles help; ctrl+f opens the filter prompt;
// c/b open argument prompts; q quits; ctrl+1/2/3 toggle the filter field.
func (m model) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.mode = modeHelp
		m.help.GotoTop()
		return m, nil
	case "ctrl+f":
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

// openFilterPrompt enters the ctrl+f prompt with the committed filter pre-filled,
// so editing starts from the current state. effectiveFilter() makes the visible
// rows track the draft live while the prompt is open.
func (m model) openFilterPrompt() (model, tea.Cmd) {
	m.mode = modeFilterPrompt
	m = m.applyPromptLabel(filterLabel)
	m.prompt.SetValue(m.filter)
	m.prompt.CursorEnd()
	return m, m.prompt.Focus()
}

// updateFilterPrompt handles keys with the filter prompt focused. Enter commits
// the draft to m.filter and closes; ctrl+f reverts (discards the draft) and
// closes; ESC clears the draft, or — when the draft is already empty — reverts
// and closes (same as ctrl+f-while-open). ctrl+1/2/3 still toggle the field;
// the modes row in the header reflects the change live.
func (m model) updateFilterPrompt(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+f":
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
		return m, nil
	case "ctrl+2":
		m.field = fieldName
		return m, nil
	case "ctrl+3":
		m.field = fieldBranch
		return m, nil
	}
	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

// openCheckoutPrompt enters the c prompt: empty draft, branch suggestions
// populated from the visible repos.
func (m model) openCheckoutPrompt() (model, tea.Cmd) {
	m.mode = modeCheckoutPrompt
	m = m.applyPromptLabel(switchBranchLabel)
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

// openBranchPrompt enters the b prompt: empty draft. Existing branches are
// surfaced as suggestions for reference (you're inventing a new name, but the
// list shows neighbours to avoid collisions); Tab cycles them in. Picking an
// existing name and pressing Enter will fail at the git layer, and the typed
// error surfaces on the row like any other failure.
func (m model) openBranchPrompt() (model, tea.Cmd) {
	m.mode = modeBranchPrompt
	m = m.applyPromptLabel(newBranchLabel)
	m.prompt.SetValue("")
	m = m.recomputeSuggestions()
	return m, m.prompt.Focus()
}

// updateBranchPrompt handles keys with the b prompt focused. Enter runs
// `checkout -b <name>` on the filtered repos and closes; ESC clears the draft
// or — when empty — reverts and closes. Tab/shift+tab cycle branch suggestions
// inline. No retrigger-close (same reason as the c prompt: `b` is a typeable
// letter).
func (m model) updateBranchPrompt(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		return m.cycleSuggestion(1), nil
	case "shift+tab":
		return m.cycleSuggestion(-1), nil
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
		m = m.recomputeSuggestions()
		return m, nil
	}
	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	m = m.recomputeSuggestions()
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

// updateHelp handles keys with the help overlay open. ? or ESC closes it; every
// other key is forwarded to the viewport for scrolling (↑/↓, j/k, PgUp/PgDn,
// Home/End). q stays unbound — you back out of help before quitting the app.
func (m model) updateHelp(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc":
		m.mode = modeList
		return m, nil
	}
	var cmd tea.Cmd
	m.help, cmd = m.help.Update(msg)
	return m, cmd
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
// fires cmdFor against each, kicking the spinner for the running rows.
func (m model) runOnFiltered(cmdFor func(name string, repo git.Repo) tea.Cmd) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, i := range m.matchedIndexes() {
		m.repos[i].cmd = cmdRunning
		m.repos[i].cmdErr = nil
		cmds = append(cmds, cmdFor(m.repos[i].name, m.repos[i].repo))
	}
	if len(cmds) == 0 {
		return m, nil
	}
	var tick tea.Cmd
	m, tick = m.kickSpinner()
	return m, tea.Batch(append(cmds, tick)...)
}

// startLoad opens a load cycle for the named repo: bumps the in-flight counter
// by the three reads about to fire and clears the cycle's loadErr so a prior
// failure doesn't linger past a fresh, successful read.
func (m model) startLoad(name string) model {
	for i := range m.repos {
		if m.repos[i].name == name {
			m.repos[i].loading += 3
			m.repos[i].loadErr = nil
			break
		}
	}
	return m
}

// loadDone records one finished read for the named repo: decrements the
// in-flight counter and, on failure, records the error as the cycle's loadErr
// (read once the counter settles to 0).
func (m model) loadDone(name string, loadErr error) model {
	for i := range m.repos {
		if m.repos[i].name == name {
			if m.repos[i].loading > 0 {
				m.repos[i].loading--
			}
			if loadErr != nil {
				m.repos[i].loadErr = loadErr
			}
			break
		}
	}
	return m
}

// clearCmdError forgets the named repo's last command outcome, so an explicit
// `r` refresh wipes a stale ✗ and its one-liner. The post-command auto-refresh
// uses startLoad directly and does not call this, so a just-failed command's
// error survives its own follow-up reads.
func (m model) clearCmdError(name string) model {
	for i := range m.repos {
		if m.repos[i].name == name {
			m.repos[i].cmd = cmdNone
			m.repos[i].cmdErr = nil
			break
		}
	}
	return m
}

// anyBusy reports whether any repo is reading or has a command in flight — the
// condition under which the spinner should keep ticking.
func (m model) anyBusy() bool {
	for i := range m.repos {
		if m.repos[i].loading > 0 || m.repos[i].cmd == cmdRunning {
			return true
		}
	}
	return false
}

// kickSpinner starts the spinner tick loop if it isn't already running. The
// spinning guard keeps concurrent commands/loads from stacking parallel tick
// chains (which would spin too fast); the loop stops itself on the first
// TickMsg seen while nothing is busy.
func (m model) kickSpinner() (model, tea.Cmd) {
	if m.spinning {
		return m, nil
	}
	m.spinning = true
	return m, m.spinner.Tick
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
		m = m.startLoad(r.name).clearCmdError(r.name)
		cmds = append(cmds, statusCmd(r.name, r.repo), diffCmd(r.name, r.repo), branchesCmd(r.name, r.repo))
	}
	if len(cmds) == 0 {
		return m, nil
	}
	var tick tea.Cmd
	m, tick = m.kickSpinner()
	return m, tea.Batch(append(cmds, tick)...)
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
		content := lipgloss.JoinVertical(lipgloss.Left, m.helpHeader(), m.help.View(), m.helpFooter())
		return tea.View{Content: content, AltScreen: true, MouseMode: tea.MouseModeCellMotion}
	}
	header := m.framedHeader(lipgloss.JoinVertical(lipgloss.Left, m.headerTop(), m.headerBottom()))
	list := m.listContent()

	if m.height > 0 {
		listHeight := m.height - 3
		if listHeight < 1 {
			listHeight = 1
		}
		listArea := lipgloss.NewStyle().Height(listHeight).Render(list)
		return tea.View{Content: lipgloss.JoinVertical(lipgloss.Left, header, listArea), AltScreen: true}
	}
	return tea.View{Content: lipgloss.JoinVertical(lipgloss.Left, header, list), AltScreen: true}
}

// framedHeader composes the top chrome shared by the list view and the help
// overlay: the given left block beside the right-corner version/PID block, with
// the full-width dim rule beneath. The left block is width-constrained so the
// right block sits flush in the corner.
func (m model) framedHeader(left string) string {
	right := m.rightBlock()
	var topBlock string
	if m.width > 0 {
		leftWidth := m.width - lipgloss.Width(right)
		if leftWidth < 0 {
			leftWidth = 0
		}
		left = lipgloss.NewStyle().Width(leftWidth).Render(left)
		topBlock = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	} else {
		topBlock = lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	}
	rule := colorDim.Render(strings.Repeat("─", m.width))
	return lipgloss.JoinVertical(lipgloss.Left, topBlock, rule)
}

// helpHeader is the help overlay's fixed top chrome: the dim "Log" label over
// the log path on the left, the shared version/PID block on the right, framed
// by the same rule as the list header.
func (m model) helpHeader() string {
	left := lipgloss.JoinVertical(lipgloss.Left, colorDim.Render("Log"), colorDim.Render(m.logPath))
	return m.framedHeader(left)
}

// helpFooter is the help overlay's fixed bottom line: a dim back hint on the
// left and the scroll percentage on the right, the percentage omitted when the
// bindings fit without scrolling.
func (m model) helpFooter() string {
	hint := colorDim.Render("? / esc: back")
	if m.help.TotalLineCount() <= m.help.VisibleLineCount() {
		return hint
	}
	pct := colorDim.Render(fmt.Sprintf("%3.f%%", m.help.ScrollPercent()*100))
	gap := m.width - lipgloss.Width(hint) - lipgloss.Width(pct)
	if gap < 1 {
		gap = 1
	}
	return hint + strings.Repeat(" ", gap) + pct
}

// rightBlock is the static right-corner chrome: "gbx <version>" over
// "PID: <pid>", both dim, right-aligned. Shown in every mode.
func (m model) rightBlock() string {
	ver := colorDim.Render("gbx " + m.version)
	pid := colorDim.Render("PID: " + strconv.Itoa(m.pid))
	return lipgloss.JoinVertical(lipgloss.Right, ver, pid)
}

// headerTop is row 1: the active prompt's input while a prompt is open, or the
// committed filter status (label + value, with dim "none" when empty) when in
// list mode.
func (m model) headerTop() string {
	switch m.mode {
	case modeFilterPrompt, modeCheckoutPrompt, modeBranchPrompt:
		return m.prompt.View()
	}
	if m.filter == "" {
		return filterLabel + colorDim.Render("none")
	}
	return filterLabel + m.filter
}

// headerBottom is row 2: the filter-field mode chips when in list or filter
// prompt mode, or the branch suggestion row when in a c/b prompt (dim
// "(no matches)" if the draft filters them all out).
func (m model) headerBottom() string {
	switch m.mode {
	case modeCheckoutPrompt, modeBranchPrompt:
		return m.suggestionLine()
	}
	return m.modesLine()
}

// modesLine renders the C-1/2/3 chips, each a dim "<C-N>" key prefix followed
// by a label — the active chip's label bold + accent, the others dim —
// separated by middle dots.
func (m model) modesLine() string {
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	chips := []struct {
		field filterField
		key   string
		label string
	}{
		{fieldNameBranch, "<C-1>", "name + branch"},
		{fieldName, "<C-2>", "name"},
		{fieldBranch, "<C-3>", "branch"},
	}
	parts := make([]string, len(chips))
	for i, c := range chips {
		labelStyle := colorDim
		if c.field == m.field {
			labelStyle = active
		}
		parts[i] = colorDim.Render(c.key+" ") + labelStyle.Render(c.label)
	}
	line := strings.Join(parts, " · ")
	if m.width > 0 {
		line = ansi.Truncate(line, m.width, "…")
	}
	return line
}

// suggestionLine renders the c/b-prompt's autocomplete options, the highlighted
// one (cycled by tab) reversed. Falls back to a dim "(no matches)" hint when
// the draft narrows the set to empty, so the row stays anchored.
func (m model) suggestionLine() string {
	if len(m.suggestions) == 0 {
		return colorDim.Render("(no matches)")
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

// gutterCell is a row's 2-wide left indicator: the dim spinner while the row is
// busy (reading or running a command), a red ✗ once it has settled with a
// command or load error, and blank otherwise (success is silent).
func (m model) gutterCell(r repoEntry) string {
	if r.loading > 0 || r.cmd == cmdRunning {
		return m.spinner.View()
	}
	if r.cmdErr != nil || r.loadErr != nil {
		return colorRed.Render("✗")
	}
	return ""
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
	nameWidth, branchWidth, trackingWidth, stateWidth, diffWidth := 0, 0, 0, 0, 0
	for _, r := range m.repos {
		if w := lipgloss.Width(r.name); w > nameWidth {
			nameWidth = w
		}
		if w := lipgloss.Width(branchText(r)); w > branchWidth {
			branchWidth = w
		}
		if w := lipgloss.Width(trackingText(r)); w > trackingWidth {
			trackingWidth = w
		}
		if w := lipgloss.Width(stateText(r)); w > stateWidth {
			stateWidth = w
		}
		if w := lipgloss.Width(diffText(r)); w > diffWidth {
			diffWidth = w
		}
	}
	gutterCol := lipgloss.NewStyle().Width(2) // spinner / ✗ slot, 1 glyph + 1 pad
	nameCol := lipgloss.NewStyle().Width(nameWidth)
	branchCol := lipgloss.NewStyle().Width(branchWidth)
	trackingCol := lipgloss.NewStyle().Width(trackingWidth)
	stateCol := lipgloss.NewStyle().Width(stateWidth)
	diffCol := lipgloss.NewStyle().Width(diffWidth)

	// Underline the filter-matched characters, scoped to the searched field:
	// C-2 lights only the name, C-3 only the branch, C-1 (default) each column
	// independently wherever it matched. Terms are parsed once for all rows.
	terms := parseTerms(m.effectiveFilter())
	hlName := m.field != fieldBranch
	hlBranch := m.field != fieldName

	rows := make([]string, len(matched))
	for i, r := range matched {
		name := r.name
		if hlName {
			name = renderHighlight(r.name, matchPositions(terms, r.name), lipgloss.NewStyle())
		}
		branch := branchText(r)
		if hlBranch && r.status != nil {
			branch = renderHighlight(r.status.branch, matchPositions(terms, r.status.branch), branchStyle(r.status.branch))
		}
		cols := []string{gutterCol.Render(m.gutterCell(r)), nameCol.Render(name), "  ", branchCol.Render(branch), "  ", trackingCol.Render(trackingText(r)), "  ", stateCol.Render(stateText(r)), "  ", diffCol.Render(diffText(r))}
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

// trackingText is the upstream-relationship column (⌀ / ahead-behind arrows),
// blank until status loads.
func trackingText(r repoEntry) string {
	if r.status == nil {
		return ""
	}
	return r.status.trackingField()
}

// diffText is the +/- line-changes column for a row: "..." until the diff
// loads, blank once it has settled with no changes (success is silent), else
// the "+N -N" aggregate.
func diffText(r repoEntry) string {
	if r.diff == nil {
		return "..."
	}
	if r.diff.empty() {
		return ""
	}
	return r.diff.String()
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
