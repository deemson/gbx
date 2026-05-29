package tui

import (
	"fmt"
	"strings"
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
	{"ctrl+f", "filter prompt"},
	{"r", "refresh filtered repos"},
	{"f", "fetch on filtered repos"},
	{"p", "pull (fast-forward) on filtered repos"},
	{"c", "Switch Branch prompt (checkout <ref>)"},
	{"b", "New Branch prompt (checkout -b <name>)"},
	{"ctrl+1", "filter field: name + branch (default)"},
	{"ctrl+2", "filter field: name"},
	{"ctrl+3", "filter field: branch"},
	{"q", "quit"},
	{"ctrl+c", "quit (any mode)"},
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

func helpContent() string {
	var b strings.Builder
	b.WriteString("gbx — keys\n\nlist mode\n\n")
	for _, kb := range listBindings {
		fmt.Fprintf(&b, "  %-10s  %s\n", kb.keys, kb.desc)
	}
	b.WriteString("\nprompts (ctrl+f filter · c Switch Branch · b New Branch)\n\n")
	for _, kb := range promptBindings {
		fmt.Fprintf(&b, "  %-10s  %s\n", kb.keys, kb.desc)
	}
	b.WriteString("\nfilter syntax (space = AND)\n\n")
	for _, fs := range filterSyntax {
		fmt.Fprintf(&b, "  %-10s  %s\n", fs.keys, fs.desc)
	}
	b.WriteString("\n? or esc: back\n")
	return b.String()
}
