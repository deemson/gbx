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
	{"r", "refresh filtered repos"},
	{"f", "fetch on filtered repos"},
	{"p", "pull (fast-forward) on filtered repos"},
	{"c", "Checkout prompt (checkout <ref>)"},
	{"b", "New Branch prompt (checkout -b <name>)"},
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

// promptBindings document the shared behavior of the ctrl+f / c / b prompts.
// ctrl+f while the filter prompt is open reverts; c and b lack that toggle
// (their letters are typeable). ctrl+1/2/3 toggle the field in the filter
// prompt only; in c/b they fall through to the textinput.
var promptBindings = []keyBinding{
	{"type", "edit the draft"},
	{"enter", "apply: ctrl+f commits filter · c runs checkout · b runs checkout -b"},
	{"esc", "clear the draft; if already empty, revert and close"},
	{"ctrl+f", "(filter prompt only) revert and close, discarding the draft"},
	{"tab", "next branch suggestion (c / b prompts)"},
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

// footerListBindings / footerFilterBindings / footerArgBindings are the curated
// one-line keybinding hints shown in the always-visible bottom footer, switched
// by mode. They carry shorter labels than the ? overlay above (a footer sheds
// hints to fit its width), so they're a separate, deliberately terse surface.
// Keys are angle-bracketed to mirror the header's <C-f>/<C-1> hint style.
var footerListBindings = []keyBinding{
	{"<C-f>", "filter"},
	{"<r>", "refresh"},
	{"<f>", "fetch"},
	{"<p>", "pull"},
	{"<c>", "checkout"},
	{"<b>", "new branch"},
	{"<?>", "help"},
	{"<q>", "quit"},
}

var footerFilterBindings = []keyBinding{
	{"<enter>", "apply"},
	{"<esc>", "clear/close"},
	{"<C-f>", "cancel"},
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
	section("actions menu (enter)", actionMenuBindings)
	b.WriteString("\n")
	section("prompts (ctrl+f filter · c Checkout · b New Branch)", promptBindings)
	b.WriteString("\n")
	section("filter syntax (space = AND)", filterSyntax)
	return b.String()
}
