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
// actions on the filtered set; F-keys open overlays; ctrl bindings toggle the
// filter field. The bottom bar shows only "F1 Help" — this overlay is where
// the rest of the keys are explained.
var listBindings = []keyBinding{
	{"F1", "toggle this help"},
	{"F4", "filter prompt"},
	{"r", "refresh filtered repos"},
	{"f", "fetch on filtered repos"},
	{"p", "pull (fast-forward) on filtered repos"},
	{"c", "checkout <ref> prompt"},
	{"b", "checkout -b <name> prompt"},
	{"ctrl+1", "filter field: name + branch (default)"},
	{"ctrl+2", "filter field: name"},
	{"ctrl+3", "filter field: branch"},
	{"q", "quit"},
	{"ctrl+c", "quit (any mode)"},
}

// promptBindings document the shared behavior of the F4 / c / b prompts. F4
// while open reverts; c and b lack that toggle (their letters are typeable).
var promptBindings = []keyBinding{
	{"type", "edit the draft"},
	{"enter", "apply: F4 commits filter · c runs checkout · b runs checkout -b"},
	{"esc", "clear the draft; if already empty, revert and close"},
	{"F4", "(F4 prompt only) revert and close, discarding the draft"},
	{"tab", "next branch suggestion (c prompt only)"},
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
	b.WriteString("\nprompts (F4 filter · c checkout · b checkout -b)\n\n")
	for _, kb := range promptBindings {
		fmt.Fprintf(&b, "  %-10s  %s\n", kb.keys, kb.desc)
	}
	b.WriteString("\nfilter syntax (space = AND)\n\n")
	for _, fs := range filterSyntax {
		fmt.Fprintf(&b, "  %-10s  %s\n", fs.keys, fs.desc)
	}
	b.WriteString("\nF1 or esc: back\n")
	return b.String()
}
