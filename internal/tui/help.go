package tui

import (
	"fmt"
	"strings"
)

// keyBinding is one row of the help overlay. This slice is the source of truth
// for the filter-mode bindings.
type keyBinding struct {
	keys string
	desc string
}

var keyBindings = []keyBinding{
	{"type", "filter repos (see filter syntax)"},
	{"ctrl+1", "filter by name + branch (default)"},
	{"ctrl+2", "filter by name"},
	{"ctrl+3", "filter by branch"},
	{"enter", "apply filter, enter command mode"},
	{"ctrl+r", "refresh the filtered repos"},
	{"ctrl+g", "toggle this help"},
	{"esc", "quit"},
	{"ctrl+c", "quit"},
}

// commandBindings document command mode, where the line runs one of the four
// supported git commands against the filtered repos with autocomplete.
var commandBindings = []keyBinding{
	{"type", "edit the command (see commands)"},
	{"tab", "next autocomplete suggestion"},
	{"shift+tab", "previous suggestion"},
	{"enter", "run on the filtered repos"},
	{"esc", "back to filter (clears it)"},
}

// commands is the fixed command vocabulary command mode accepts.
var commands = []keyBinding{
	{"checkout <ref>", "switch to an existing branch"},
	{"checkout -b <name>", "create a new branch"},
	{"fetch", "fetch from the remote"},
	{"pull", "fast-forward pull"},
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
	b.WriteString("gbx — keys\n\nfilter mode\n\n")
	for _, kb := range keyBindings {
		fmt.Fprintf(&b, "  %-10s  %s\n", kb.keys, kb.desc)
	}
	b.WriteString("\ncommand mode\n\n")
	for _, kb := range commandBindings {
		fmt.Fprintf(&b, "  %-10s  %s\n", kb.keys, kb.desc)
	}
	b.WriteString("\ncommands\n\n")
	for _, c := range commands {
		fmt.Fprintf(&b, "  %-18s  %s\n", c.keys, c.desc)
	}
	b.WriteString("\nfilter syntax (space = AND)\n\n")
	for _, fs := range filterSyntax {
		fmt.Fprintf(&b, "  %-10s  %s\n", fs.keys, fs.desc)
	}
	b.WriteString("\nesc: back\n")
	return b.String()
}
