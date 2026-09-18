package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// keyBinding is one row of the help overlay. These slices are the source of
// truth for the bindings — the overlay is the user-facing discovery surface.
type keyBinding struct {
	keys string
	desc string
}

type iconBinding struct {
	icon string
	desc string
}

// listBindings document the default (list) mode: letter keys dispatch git
// actions on the filtered set; ? toggles help; ctrl+f opens the filter prompt;
// ctrl+1/2/3 toggle the filter field. The header is always visible and shows
// the committed filter + active field — this overlay is where the rest of the
// keys are explained.
var listBindings = []keyBinding{
	{"?", "toggle this help"},
	{"↑/↓ ctrl+p/ctrl+n", "move the cursor"},
	{"ctrl+u/ctrl+d", "half-page up/down"},
	{"enter", "open the actions menu for the cursored repo"},
	{"ctrl+f", "filter prompt"},
	{"ctrl+s", "search prompt (jump cursor to matches)"},
	{"r", "refresh filtered repos"},
	{"f", "fetch on filtered repos"},
	{"p", "pull (fast-forward) on filtered repos"},
	{"s", "Switch prompt (switch <branch>)"},
	{"b", "New Branch prompt (switch -c <name>)"},
	{"ctrl+1", "filter field: name + branch (default)"},
	{"ctrl+2", "filter field: name"},
	{"ctrl+3", "filter field: branch"},
	{"q", "quit"},
	{"ctrl+c", "quit (any mode)"},
}

// actionMenuBindings document the enter-key digit menu: each configured action
// is bound to its 1-based digit and runs in the cursored repo's directory,
// suspending gbx until the launched tool exits.
var actionMenuBindings = []keyBinding{
	{"1-9", "run that action in the cursored repo's directory"},
	{"esc / enter / q", "close without running anything"},
}

// promptBindings document the shared behavior of the ctrl+f / ctrl+s / s / b
// prompts. ctrl+f while the filter prompt is open reverts; s and b lack that
// toggle (their letters are typeable). ctrl+1/2/3 toggle the field in the filter
// and search prompts; in s/b they fall through to the textinput. The search
// prompt moves the cursor instead of narrowing: enter keeps the landing spot,
// esc / ctrl+s cancel and restore the cursor, ↓/↑ (or ctrl+n/ctrl+p) walk the
// matches.
var promptBindings = []keyBinding{
	{"type", "edit the draft"},
	{"enter", "apply: ctrl+f commits filter · ctrl+s keeps cursor · s runs switch · b runs switch -c"},
	{"esc", "clear the draft (ctrl+s: restore cursor); if already empty, revert and close"},
	{"ctrl+f", "(filter prompt only) revert and close, discarding the draft"},
	{"↓/↑ ctrl+n/ctrl+p", "(search prompt) next / previous match"},
	{"tab", "next branch suggestion (s / b prompts)"},
	{"shift+tab", "previous suggestion"},
}

// filterSyntax documents the fzf-style filter DSL: space-separated terms ANDed
// together, each fuzzy by default unless anchored or negated.
var filterSyntax = []keyBinding{
	{"foo", "fuzzy match"},
	{"^foo", "starts with foo"},
	{"foo$", "ends with foo"},
	{"!foo", "exclude foo"},
}

// iconBindings explain every symbolic status shown in a repository row. Each
// sample uses the same style as its list-view counterpart.
var iconBindings = []iconBinding{
	{colorGreen.Render("✓"), "command succeeded"},
	{colorRed.Render("✗"), "load or command failed"},
	{colorDim.Render("⌀"), "branch has no upstream"},
	{colorCyan.Render("↑"), "commits ahead of upstream"},
	{colorCyan.Render("↓"), "commits behind upstream"},
	{colorYellow.Render("~"), "modified files"},
	{colorGreen.Render("✚"), "added files"},
	{colorRed.Render("✖"), "deleted files"},
	{colorMagenta.Render("»"), "renamed files"},
	{colorDim.Render("…"), "untracked files"},
	{colorCyan.Render("≡"), "stashes"},
	{colorBrightRed.Render("‡"), "conflicted files"},
	{colorGreen.Render("+"), "added lines"},
	{colorRed.Render("-"), "deleted lines"},
}

// footerListBindings / footerFilterBindings / footerArgBindings are the curated
// one-line keybinding hints shown in the always-visible bottom footer, switched
// by mode. They carry shorter labels than the ? overlay above (a footer sheds
// hints to fit its width), so they're a separate, deliberately terse surface.
// Keys are angle-bracketed to mirror the header's <C-f>/<C-1> hint style.
var footerListBindings = []keyBinding{
	{"<C-f>", "filter"},
	{"<C-s>", "search"},
	{"<r>", "refresh"},
	{"<f>", "fetch"},
	{"<p>", "pull"},
	{"<s>", "switch"},
	{"<b>", "new branch"},
	{"<?>", "help"},
	{"<q>", "quit"},
}

var footerFilterBindings = []keyBinding{
	{"<enter>", "apply"},
	{"<esc>", "clear/close"},
	{"<C-f>", "cancel"},
}

var footerSearchBindings = []keyBinding{
	{"<enter>", "keep"},
	{"<esc/C-s>", "cancel"},
	{"<↓/↑>", "next/prev"},
}

var footerArgBindings = []keyBinding{
	{"<enter>", "apply"},
	{"<esc>", "clear/close"},
	{"<tab/S-tab>", "cycle"},
}

var footerActionBindings = []keyBinding{
	{"<1-9>", "run"},
	{"<esc>", "cancel"},
}

// helpHeading styles a section header — cyan bold, the app's accent (the active
// filter chip). helpKey styles the key column yellow so keys pop against the
// default-foreground descriptions.
var (
	helpHeading = colorCyan.Bold(true)
	helpKey     = colorYellow
)

// helpContent is the scrollable body of the help overlay: the three binding
// sections, no title and no back hint (those live in the fixed header/footer).
// Keys are colored and padded to the section's widest key; descriptions stay
// default.
func helpContent() string {
	var b strings.Builder
	section := func(title string, bindings []keyBinding) {
		b.WriteString(helpHeading.Render(title))
		b.WriteString("\n\n")
		keyCol := 0
		for _, kb := range bindings {
			keyCol = max(keyCol, lipgloss.Width(kb.keys))
		}
		for _, kb := range bindings {
			pad := keyCol - lipgloss.Width(kb.keys)
			fmt.Fprintf(&b, "  %s%s  %s\n", helpKey.Render(kb.keys), strings.Repeat(" ", pad), kb.desc)
		}
	}
	section("list mode", listBindings)
	b.WriteString("\n")
	b.WriteString(helpHeading.Render("repository row icons"))
	b.WriteString("\n\n")
	iconCol := 0
	for _, ib := range iconBindings {
		iconCol = max(iconCol, lipgloss.Width(ib.icon))
	}
	for _, ib := range iconBindings {
		pad := iconCol - lipgloss.Width(ib.icon)
		fmt.Fprintf(&b, "  %s%s  %s\n", ib.icon, strings.Repeat(" ", pad), ib.desc)
	}
	b.WriteString("\n")
	section("actions menu (enter)", actionMenuBindings)
	b.WriteString("\n")
	section("prompts (ctrl+f filter · ctrl+s search · s Switch · b New Branch)", promptBindings)
	b.WriteString("\n")
	section("filter syntax (space = AND)", filterSyntax)
	return b.String()
}
